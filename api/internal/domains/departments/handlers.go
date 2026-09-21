package departments

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog"
)

type handler struct {
	departmentsService *DepartmentsService
	logger             zerolog.Logger
}

func NewHandler(departmentsService *DepartmentsService, logger zerolog.Logger) *handler {
	return &handler{
		departmentsService: departmentsService,
		logger:             logger.With().Str("handler", "DepartmentsHandler").Str("domain", "departments").Logger(),
	}
}

func (h *handler) listDepartments(
	ctx context.Context,
	input *struct{},
) (*ListDepartmentsResponse, error) {
	departments, err := h.departmentsService.List(ctx)
	if err != nil {
		return nil, h.departmentsHTTPError(err, "Failed to list departments")
	}
	return &ListDepartmentsResponse{
		Body: departments,
	}, nil
}

func (h *handler) departmentsHTTPError(err error, fallback string) error {
	h.logger.Warn().Err(err).Msg(fallback)
	return huma.Error500InternalServerError(fallback)
}
