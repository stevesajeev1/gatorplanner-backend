package classes

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog"
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
	},
) (*SearchClassesResponse, error) {
	classes, err := h.classesService.Search(ctx, input.TermID, &input.Body)
	if err != nil {
		return nil, classesHTTPError(err, "Failed to search classes")
	}
	return &SearchClassesResponse{Body: classes}, nil
}

func classesHTTPError(err error, fallback string) error {
	if errors.Is(err, ErrSearchClassesSearchOrFilterRequired) {
		return huma.Error400BadRequest(err.Error())
	}

	return huma.Error500InternalServerError(fallback)
}
