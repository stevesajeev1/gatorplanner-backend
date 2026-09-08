package database

import (
	"context"

	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
)

type CoursesRepository struct {
	db *dependencies.DB
}

func NewCoursesRepository(db *dependencies.DB) *CoursesRepository {
	return &CoursesRepository{db: db}
}

func (r *CoursesRepository) ListCoursesByID(
	ctx context.Context,
	ids []int32,
) ([]sqlc.ListCoursesByIDRow, error) {
	return r.db.Query.ListCoursesByID(ctx, ids)
}
