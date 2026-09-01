package classes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterRoutes(handler *handler, group huma.API) {
	huma.Register(group, huma.Operation{
		OperationID:   "search-classes",
		Method:        http.MethodPost,
		Summary:       "Search Classes",
		Description:   "Searches classes, with optional filtering.",
		Tags:          []string{"Classes"},
		Path:          "/search",
		Errors:        []int{http.StatusInternalServerError},
		DefaultStatus: http.StatusOK,
	}, handler.searchClasses)
}