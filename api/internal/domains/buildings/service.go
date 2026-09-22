package buildings

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/database"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
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
) ([]sqlc.TypedListBuildingsRow, error) {
	rawBuildings, err := s.buildingsDBRepo.ListBuildings(ctx)
	if err != nil {
		return nil, err
	}

	buildings, err := sqlc.RawListBuildingsRows(rawBuildings).Typed()
	if err != nil {
		return nil, err
	}
	return buildings, nil
}
