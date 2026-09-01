package classes

import (
	"context"
	"net/http"

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
	input *struct{
		TermID string `path:"termID"`
	},
) (*SearchClassesOutput, error) {
	return &SearchClassesOutput{Status: http.StatusOK}, nil
}