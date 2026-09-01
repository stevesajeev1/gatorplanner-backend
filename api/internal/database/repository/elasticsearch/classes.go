package elasticsearch

import "github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"

type ClassesESRepository struct {
	es *dependencies.ES
}

func NewClassesESRepository(es *dependencies.ES) *ClassesESRepository {
	return &ClassesESRepository{es: es}
}