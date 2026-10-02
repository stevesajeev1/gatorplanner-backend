package buildings

import "github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"

type ListBuildingsResponse struct {
	Body []sqlc.TypedBuilding `nullable:"false"`
}
