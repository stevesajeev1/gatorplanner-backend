package buildings

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(handler *handler, group huma.API) {
	huma.Register(group, huma.Operation{
		OperationID:   "list-buildings",
		Method:        http.MethodGet,
		Summary:       "List Buildings",
		Description:   "Lists buildings in alphabetical order.",
		Tags:          []string{"Info"},
		Path:          "/buildings",
		Errors:        []int{http.StatusInternalServerError},
		DefaultStatus: http.StatusOK,
	}, handler.listBuildings)
}
