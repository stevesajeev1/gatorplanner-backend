package buildings

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/database"
)

type BuildingsService struct {
	buildingsDBRepo *database.BuildingsRepository
	logger          zerolog.Logger
}

func NewService(
	buildingsDBRepo *database.BuildingsRepository,
	logger zerolog.Logger,
) *BuildingsService {
	return &BuildingsService{
		buildingsDBRepo: buildingsDBRepo,
		logger:          logger.With().Str("service", "BuildingsService").Str("domain", "buildings").Logger(),
	}
}

func (s *BuildingsService) List(
	ctx context.Context,
) ([]string, error) {
	return s.buildingsDBRepo.ListBuildings(ctx)
}
