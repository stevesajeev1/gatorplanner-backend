package departments

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/database"
)

type DepartmentsService struct {
	departmentsDBRepo *database.DepartmentsRepository
	logger            zerolog.Logger
}

func NewService(
	departmentsDBRepo *database.DepartmentsRepository,
	logger zerolog.Logger,
) *DepartmentsService {
	return &DepartmentsService{
		departmentsDBRepo: departmentsDBRepo,
		logger:            logger.With().Str("service", "DepartmentsService").Str("domain", "departments").Logger(),
	}
}

func (s *DepartmentsService) List(
	ctx context.Context,
) ([]string, error) {
	return s.departmentsDBRepo.ListDepartments(ctx)
}
