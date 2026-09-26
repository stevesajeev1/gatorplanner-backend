package schedules

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	schedulerv1 "github.com/stevesajeev1/gatorplanner-backend/generated/scheduler/v1"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/database"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/optimizer"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
)

var (
	ErrInvalidClasses = errors.New("invalid classes")
)

type SchedulesService struct {
	classesDBRepo          *database.ClassesRepository
	coursesDBRepo          *database.CoursesRepository
	schedulesOptimizerRepo *optimizer.SchedulesOptimizerRepository
	logger                 zerolog.Logger
}

func NewService(
	classesDBRepo *database.ClassesRepository,
	coursesDBRepo *database.CoursesRepository,
	schedulesOptimizerRepo *optimizer.SchedulesOptimizerRepository,
	logger zerolog.Logger,
) *SchedulesService {
	return &SchedulesService{
		classesDBRepo:          classesDBRepo,
		coursesDBRepo:          coursesDBRepo,
		schedulesOptimizerRepo: schedulesOptimizerRepo,
		logger:                 logger.With().Str("service", "SchedulesService").Str("domain", "schedules").Logger(),
	}
}

func (s *SchedulesService) Generate(
	ctx context.Context,
	termID int32,
	request *GenerateSchedulesRequest,
	limit uint,
	offset uint,
) (*GenerateSchedulesOutput, error) {
	classIDs := []uuid.UUID{}
	for _, choice := range request.ClassChoices {
		classIDs = append(classIDs, choice.ClassIDs...)
	}

	// Verify all class IDs exist for term
	validClassIDs, err := s.classesDBRepo.ValidateClassesForTerm(ctx, termID, classIDs)
	if err != nil {
		return nil, err
	}
	valid := make(map[uuid.UUID]struct{}, len(validClassIDs))
	for _, id := range validClassIDs {
		valid[id] = struct{}{}
	}
	invalidClasses := []uuid.UUID{}
	for _, id := range classIDs {
		if _, ok := valid[id]; !ok {
			invalidClasses = append(invalidClasses, id)
		}
	}
	if len(invalidClasses) > 0 {
		return nil, fmt.Errorf(
			"%w: classes %v do not exist for term %d",
			ErrInvalidClasses,
			invalidClasses,
			termID,
		)
	}

	// Get required information for schedule generation
	rawClasses, err := s.classesDBRepo.ListClassesForScheduler(ctx, termID, validClassIDs)
	if err != nil {
		return nil, err
	}

	classes, err := sqlc.RawListClassesForSchedulerRows(rawClasses).Typed()
	if err != nil {
		return nil, err
	}

	// Convert SortBy
	var sortBy *schedulerv1.SortBy = nil
	if request.SortBy != nil {
		for key, value := range SortByMap {
			if *request.SortBy == value {
				sortBy = &key
				break
			}
		}
	}

	generated, err := s.schedulesOptimizerRepo.GenerateSchedules(ctx, classes, sortBy, request.DayRestrictions, limit, offset)
	if err != nil {
		return nil, err
	}

	schedules := make([]Schedule, len(generated.Schedules))
	for i, schedule := range generated.Schedules {
		classes := make([]SelectedClass, len(schedule.Classes))
		for j, class := range schedule.Classes {
			classes[j] = SelectedClass{
				CourseID: uuid.MustParse(class.CourseId),
				ClassID:  uuid.MustParse(class.ClassId),
			}
		}
		schedules[i] = Schedule{
			Classes: classes,
		}
	}

	return &GenerateSchedulesOutput{
		Total:     uint(generated.Total),
		Schedules: schedules,
	}, nil
}
