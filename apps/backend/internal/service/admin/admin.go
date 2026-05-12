package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type Service struct {
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
}

func New(userRepo repository.UserRepository, adminRepo repository.AdminRepository) *Service {
	return &Service{userRepo: userRepo, adminRepo: adminRepo}
}

func (s *Service) ListUsers(ctx context.Context, f repository.UserAdminFilter) ([]*domain.User, int, error) {
	return s.userRepo.List(ctx, f)
}

func (s *Service) GetStats(ctx context.Context) (*domain.PlatformStats, error) {
	return s.adminRepo.GetStats(ctx)
}

func (s *Service) ListTestsWithAttempts(ctx context.Context, page, limit int) ([]*domain.TestWithAttempts, int, error) {
	return s.adminRepo.ListTestsWithAttempts(ctx, page, limit)
}

func (s *Service) ApproveContributor(ctx context.Context, userID uuid.UUID) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Role != domain.RoleContributor {
		return fmt.Errorf("user is not a contributor: %w", apierr.ErrValidation)
	}
	if u.Status != domain.StatusPending {
		return fmt.Errorf("user is not pending: %w", apierr.ErrValidation)
	}
	return s.userRepo.UpdateStatus(ctx, userID, domain.StatusActive)
}

func (s *Service) RejectContributor(ctx context.Context, userID uuid.UUID) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Role != domain.RoleContributor {
		return fmt.Errorf("user is not a contributor: %w", apierr.ErrValidation)
	}
	if u.Status != domain.StatusPending {
		return fmt.Errorf("user is not pending: %w", apierr.ErrValidation)
	}
	return s.userRepo.UpdateStatus(ctx, userID, domain.StatusSuspended)
}

func (s *Service) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status domain.Status) error {
	switch status {
	case domain.StatusActive, domain.StatusSuspended, domain.StatusPending:
	default:
		return fmt.Errorf("invalid status %q: %w", status, apierr.ErrValidation)
	}
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return apierr.ErrNotFound
		}
		return err
	}
	return s.userRepo.UpdateStatus(ctx, userID, status)
}
