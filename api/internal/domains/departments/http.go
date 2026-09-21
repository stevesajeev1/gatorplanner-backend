package departments

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(handler *handler, group huma.API) {
	huma.Register(group, huma.Operation{
		OperationID:   "list-departments",
		Method:        http.MethodGet,
		Summary:       "List Departments",
		Description:   "Lists departments in alphabetical order.",
		Tags:          []string{"Info"},
		Path:          "/departments",
		Errors:        []int{http.StatusInternalServerError},
		DefaultStatus: http.StatusOK,
	}, handler.listDepartments)
}
