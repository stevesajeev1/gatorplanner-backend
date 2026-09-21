package database

import (
	"context"

	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
)

type DepartmentsRepository struct {
	db *dependencies.DB
}

func NewDepartmentsRepository(db *dependencies.DB) *DepartmentsRepository {
	return &DepartmentsRepository{db: db}
}

func (r *DepartmentsRepository) ListDepartments(
	ctx context.Context,
) ([]string, error) {
	return r.db.Query.ListDepartments(ctx)
}
