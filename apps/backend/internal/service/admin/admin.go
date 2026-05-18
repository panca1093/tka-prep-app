package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

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

func (s *Service) ListTestsWithAttempts(ctx context.Context, page, limit int, educationLevel *domain.EducationLevel, categoryID *uuid.UUID) ([]*domain.TestWithAttempts, int, error) {
	return s.adminRepo.ListTestsWithAttempts(ctx, page, limit, educationLevel, categoryID)
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

func (s *Service) CreateContributor(ctx context.Context, name, email, password string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" || utf8.RuneCountInString(name) > 255 {
		return nil, fmt.Errorf("name is required (max 255 chars): %w", apierr.ErrValidation)
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("valid email is required: %w", apierr.ErrValidation)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters: %w", apierr.ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleContributor,
		Status:       domain.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, apierr.ErrConflict) {
			return nil, apierr.ErrConflict
		}
		return nil, fmt.Errorf("create contributor: %w", err)
	}
	return user, nil
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
