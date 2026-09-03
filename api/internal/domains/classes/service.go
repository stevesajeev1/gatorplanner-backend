package classes

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/elasticsearch"
)

var (
	ErrSearchClassesSearchOrFilterRequired = errors.New("Search must provide either search term(s) and/or filters")
)

type ClassesService struct {
	classesESRepo *elasticsearch.ClassesESRepository
	logger        zerolog.Logger
}

func NewService(
	classesESRepo *elasticsearch.ClassesESRepository,
	logger zerolog.Logger,
) *ClassesService {
	return &ClassesService{
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
) (*elasticsearch.SearchClassResult, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	search, err := s.classesESRepo.Search(ctx, termID, request.Search, request.Filter, limit, offset)
	if err != nil {
		return nil, err
	}

	return search, nil
}
