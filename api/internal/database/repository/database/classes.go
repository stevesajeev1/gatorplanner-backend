package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
)

type ClassesRepository struct {
	db *dependencies.DB
}

func NewClassesRepository(db *dependencies.DB) *ClassesRepository {
	return &ClassesRepository{db: db}
}

func (r *ClassesRepository) ListClassesByID(
	ctx context.Context,
	ids []uuid.UUID,
) ([]sqlc.ListClassesByIDRow, error) {
	return r.db.Query.ListClassesByID(ctx, ids)
}

func (r *ClassesRepository) ValidateClassesForTerm(
	ctx context.Context,
	termID int32,
	ids []uuid.UUID,
) ([]uuid.UUID, error) {
	return r.db.Query.ValidateClassesForTerm(ctx, sqlc.ValidateClassesForTermParams{
		TermID:   termID,
		ClassIds: ids,
	})
}

func (r *ClassesRepository) ListClassesForScheduler(
	ctx context.Context,
	termID int32,
	ids []uuid.UUID,
) ([]sqlc.ListClassesForSchedulerRow, error) {
	return r.db.Query.ListClassesForScheduler(ctx, sqlc.ListClassesForSchedulerParams{
		TermID:   termID,
		ClassIds: ids,
	})
}
