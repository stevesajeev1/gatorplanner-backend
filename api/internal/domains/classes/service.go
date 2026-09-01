package classes

import (
	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/elasticsearch"
)

type ClassesService struct {
	classesESRepo *elasticsearch.ClassesESRepository
	logger zerolog.Logger
}

func NewService(
	classesESRepo *elasticsearch.ClassesESRepository,
	logger zerolog.Logger,
) *ClassesService {
	return &ClassesService{
		classesESRepo: classesESRepo,
		logger: logger.With().Str("service", "ClassesService").Str("domain", "classes").Logger(),
	}
}