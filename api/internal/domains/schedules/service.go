package schedules

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/database"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/optimizer"
)

var (
	ErrCourseMustProvideChoices = errors.New("course must provide one or more class choices")
	ErrInvalidCourses           = errors.New("invalid courses")
	ErrInvalidClasses           = errors.New("invalid classes")
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
) ([]Schedule, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	classIDs := []uuid.UUID{}
	for _, course := range request.FixedCourses {
		classIDs = append(classIDs, course.ClassID)
	}
	for _, choice := range request.CourseChoices {
		for _, classID := range choice.ClassIDs {
			classIDs = append(classIDs, classID)
		}
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
	classes, err := s.classesDBRepo.ListClassesForScheduler(ctx, termID, validClassIDs)
	if err != nil {
		return nil, err
	}

	generated, err := s.schedulesOptimizerRepo.GenerateSchedules(ctx, classes)
	if err != nil {
		return nil, err
	}

	schedules := make([]Schedule, len(generated.Schedules))
	for i, schedule := range generated.Schedules {
		courses := make([]FixedCourse, len(schedule.Classes))
		for j, class := range schedule.Classes {
			courses[j] = FixedCourse{
				CourseID: class.CourseId,
				ClassID:  uuid.MustParse(class.ClassId),
			}
		}
		schedules[i] = Schedule{
			Courses: courses,
		}
	}

	return schedules, nil
}
