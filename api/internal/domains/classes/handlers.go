package classes

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/shared"
)

type handler struct {
	classesService *ClassesService
	logger         zerolog.Logger
}

func NewHandler(classesService *ClassesService, logger zerolog.Logger) *handler {
	return &handler{
		classesService: classesService,
		logger:         logger.With().Str("handler", "ClassesHandler").Str("domain", "classes").Logger(),
	}
}

func (h *handler) searchClasses(
	ctx context.Context,
	input *struct {
		TermID int64 `path:"termID"`
		Body   SearchClassesRequest
		shared.PaginationParams
	},
) (*SearchClassesResponse, error) {
	search, err := h.classesService.Search(ctx, input.TermID, &input.Body, input.Limit, input.Offset)
	if err != nil {
		return nil, h.classesHTTPError(err, "Failed to search classes")
	}
	return &SearchClassesResponse{
		Body: shared.Paginate(search.Items, search.Total, input.Limit, input.Offset),
	}, nil
}

func (h *handler) classesHTTPError(err error, fallback string) error {
	if errors.Is(err, ErrSearchClassesSearchOrFilterRequired) {
		return huma.Error400BadRequest(err.Error())
	}

	h.logger.Warn().Err(err).Msg(fallback)
	return huma.Error500InternalServerError(fallback)
}
