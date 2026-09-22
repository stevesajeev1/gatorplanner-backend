package database

import (
	"context"

	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
)

type BuildingsRepository struct {
	db *dependencies.DB
}

func NewBuildingsRepository(db *dependencies.DB) *BuildingsRepository {
	return &BuildingsRepository{db: db}
}

func (r *BuildingsRepository) ListBuildings(
	ctx context.Context,
) ([]sqlc.ListBuildingsRow, error) {
	return r.db.Query.ListBuildings(ctx)
}
