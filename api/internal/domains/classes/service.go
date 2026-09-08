package classes

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/database"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/elasticsearch"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
)

var (
	ErrSearchClassesSearchOrFilterRequired = errors.New("Search must provide either search term(s) and/or filters")
)

type ClassesService struct {
	coursesDBRepo *database.CoursesRepository
	classesDBRepo *database.ClassesRepository
	classesESRepo *elasticsearch.ClassesESRepository
	logger        zerolog.Logger
}

func NewService(
	coursesDBRepo *database.CoursesRepository,
	classesDBRepo *database.ClassesRepository,
	classesESRepo *elasticsearch.ClassesESRepository,
	logger zerolog.Logger,
) *ClassesService {
	return &ClassesService{
		coursesDBRepo: coursesDBRepo,
		classesDBRepo: classesDBRepo,
		classesESRepo: classesESRepo,
		logger:        logger.With().Str("service", "ClassesService").Str("domain", "classes").Logger(),
	}
}

func (s *ClassesService) Search(
	ctx context.Context,
	termID int64,
	request *SearchClassesRequest,
	limit uint,
	offset uint,
) (*SearchClassResponseOutput, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	search, err := s.classesESRepo.Search(ctx, termID, request.Search, request.Filter, limit, offset)
	if err != nil {
		return nil, err
	}

	// Flatten results
	courseIDs := make([]int32, len(search.Items))
	classIDs := []uuid.UUID{}
	for i, item := range search.Items {
		courseIDs[i] = item.CourseID

		for _, classID := range item.ClassIDs {
			classIDs = append(classIDs, classID)
		}
	}

	// Get actual documents from database
	rawCourses, err := s.coursesDBRepo.ListCoursesByID(ctx, courseIDs)
	if err != nil {
		return nil, err
	}
	rawClasses, err := s.classesDBRepo.ListClassesByID(ctx, classIDs)
	if err != nil {
		return nil, err
	}

	courses, err := sqlc.RawListCoursesByIDRows(rawCourses).Typed()
	if err != nil {
		return nil, err
	}
	classes, err := sqlc.RawListClassesByIDRows(rawClasses).Typed()
	if err != nil {
		return nil, err
	}

	// Construct items
	items := make([]SearchClassResponseItem, len(courses))
	accumulator := 0
	for i, course := range courses {
		classCount := len(search.Items[i].ClassIDs)

		items[i] = SearchClassResponseItem{
			Course:  course,
			Classes: classes[accumulator:(accumulator + classCount)],
		}
		accumulator += classCount
	}
	return &SearchClassResponseOutput{
		Total: search.Total,
		Items: items,
	}, nil
}
