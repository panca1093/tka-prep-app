package category

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type Service struct {
	repo repository.CategoryRepository
}

func New(repo repository.CategoryRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, name string) (*domain.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("category name is required: %w", apierr.ErrValidation)
	}
	return s.repo.Create(ctx, name)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, name string) (*domain.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("category name is required: %w", apierr.ErrValidation)
	}
	return s.repo.Update(ctx, id, name)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := s.repo.TestCount(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("kategori memiliki %d ujian — tidak dapat dihapus: %w", n, apierr.ErrConflict)
	}
	return s.repo.Delete(ctx, id)
}
