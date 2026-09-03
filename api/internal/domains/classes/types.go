package classes

import (
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/elasticsearch"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/classes/search"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/shared"
)

type SearchClassesRequest struct {
	Search *string        `json:"search,omitempty" minLength:"1"`
	Filter *search.Filter `json:"filter,omitempty"`
}

func (r *SearchClassesRequest) Validate() error {
	if r.Search == nil && r.Filter == nil {
		return ErrSearchClassesSearchOrFilterRequired
	}

	return nil
}

type SearchClassesResponse struct {
	Body shared.PaginatedResponse[elasticsearch.SearchClassResultItem]
}
