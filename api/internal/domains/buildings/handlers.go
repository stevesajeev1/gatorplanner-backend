package buildings

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog"
)

type handler struct {
	buildingsService *BuildingsService
	logger           zerolog.Logger
}

func NewHandler(buildingsService *BuildingsService, logger zerolog.Logger) *handler {
	return &handler{
		buildingsService: buildingsService,
		logger:           logger.With().Str("handler", "BuildingsHandler").Str("domain", "buildings").Logger(),
	}
}

func (h *handler) listBuildings(
	ctx context.Context,
	input *struct{},
) (*ListBuildingsResponse, error) {
	buildings, err := h.buildingsService.List(ctx)
	if err != nil {
		return nil, h.buildingsHTTPError(err, "Failed to list buildings")
	}
	return &ListBuildingsResponse{
		Body: buildings,
	}, nil
}

func (h *handler) buildingsHTTPError(err error, fallback string) error {
	h.logger.Warn().Err(err).Msg(fallback)
	return huma.Error500InternalServerError(fallback)
}
