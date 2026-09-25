package schedules

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(handler *handler, group huma.API) {
	huma.Register(group, huma.Operation{
		OperationID:   "generate-schedules",
		Method:        http.MethodPost,
		Summary:       "Generate Schedules",
		Description:   "Generates schedules for provided classes.",
		Tags:          []string{"Schedules"},
		Path:          "/generate",
		Errors:        []int{http.StatusInternalServerError},
		DefaultStatus: http.StatusOK,
	}, handler.generateSchedules)
}
