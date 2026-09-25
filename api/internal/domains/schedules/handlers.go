package schedules

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/shared"
)

type handler struct {
	schedulesService *SchedulesService
	logger           zerolog.Logger
}

func NewHandler(schedulesService *SchedulesService, logger zerolog.Logger) *handler {
	return &handler{
		schedulesService: schedulesService,
		logger:           logger.With().Str("handler", "SchedulesHandler").Str("domain", "schedules").Logger(),
	}
}

func (h *handler) generateSchedules(
	ctx context.Context,
	input *struct {
		TermID int32 `path:"termID"`
		Body   GenerateSchedulesRequest
		shared.PaginationParams
	},
) (*GenerateSchedulesResponse, error) {
	schedules, err := h.schedulesService.Generate(ctx, input.TermID, &input.Body, input.Limit, input.Offset)
	if err != nil {
		return nil, h.schedulesHTTPError(err, "Failed to generate schedules")
	}
	return &GenerateSchedulesResponse{
		Body: shared.Paginate(schedules, uint(len(schedules)), input.Limit, input.Offset),
	}, nil
}

func (h *handler) schedulesHTTPError(err error, fallback string) error {
	if errors.Is(err, ErrInvalidClasses) {
		return huma.Error400BadRequest(err.Error())
	}

	h.logger.Warn().Err(err).Msg(fallback)
	return huma.Error500InternalServerError(fallback)
}
