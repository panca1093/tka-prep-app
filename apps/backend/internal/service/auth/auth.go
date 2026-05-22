package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type Config struct {
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

type Service struct {
	cfg    Config
	users  repository.UserRepository
	tokens repository.RefreshTokenRepository
}

func NewService(cfg Config, users repository.UserRepository, tokens repository.RefreshTokenRepository) *Service {
	return &Service{cfg: cfg, users: users, tokens: tokens}
}

type RegisterInput struct {
	Name           string
	Email          string
	Password       string
	Role           domain.Role
	EducationLevel *domain.EducationLevel
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*domain.User, TokenPair, error) {
	if err := validateRegisterInput(in); err != nil {
		return nil, TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	status := domain.StatusActive
	if in.Role == domain.RoleContributor {
		status = domain.StatusPending
	}

	var el *domain.EducationLevel
	if in.Role == domain.RoleStudent && in.EducationLevel != nil {
		el = in.EducationLevel
	}

	user := &domain.User{
		ID:             uuid.New(),
		Name:           strings.TrimSpace(in.Name),
		Email:          strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash:   string(hash),
		Role:           in.Role,
		Status:         status,
		EducationLevel: el,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, apierr.ErrConflict) {
			return nil, TokenPair{}, apierr.ErrConflict
		}
		return nil, TokenPair{}, fmt.Errorf("create user: %w", err)
	}

	// Pending contributors must wait for admin approval before they can log in.
	if user.Status == domain.StatusPending {
		return user, TokenPair{}, nil
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*domain.User, TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return nil, TokenPair{}, apierr.ErrUnauthorized
		}
		return nil, TokenPair{}, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, TokenPair{}, apierr.ErrUnauthorized
	}

	if user.Status == domain.StatusPending {
		return nil, TokenPair{}, apierr.ErrPending
	}
	if user.Status == domain.StatusSuspended {
		return nil, TokenPair{}, apierr.ErrForbidden
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (*domain.User, TokenPair, error) {
	hash := hashToken(rawToken)

	storedToken, err := s.tokens.FindByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return nil, TokenPair{}, apierr.ErrUnauthorized
		}
		return nil, TokenPair{}, fmt.Errorf("find refresh token: %w", err)
	}

	if time.Now().After(storedToken.ExpiresAt) {
		_ = s.tokens.DeleteByTokenHash(ctx, hash)
		return nil, TokenPair{}, apierr.ErrUnauthorized
	}

	// Rotate: delete old token before issuing new one (detects reuse theft)
	if err := s.tokens.DeleteByTokenHash(ctx, hash); err != nil {
		return nil, TokenPair{}, fmt.Errorf("rotate token: %w", err)
	}

	user, err := s.users.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, TokenPair{}, fmt.Errorf("find user for refresh: %w", err)
	}

	if user.Status == domain.StatusPending || user.Status == domain.StatusSuspended {
		return nil, TokenPair{}, apierr.ErrUnauthorized
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	if err := s.tokens.DeleteByTokenHash(ctx, hash); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

type UpdateProfileInput struct {
	EducationLevel *domain.EducationLevel
	Gender         *domain.Gender
	Phone          *string
	AvatarURL      *string
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find user for profile update: %w", err)
	}

	// Education level only applies to students.
	if user.Role != domain.RoleStudent && in.EducationLevel != nil {
		return nil, fmt.Errorf("%w: education_level is only applicable for student role", apierr.ErrValidation)
	}

	// Education level can only be set once. Once a student has a level, it is frozen.
	if in.EducationLevel != nil && user.EducationLevel != nil {
		return nil, fmt.Errorf("%w: education level can only be set when empty — contact support to change", apierr.ErrValidation)
	}

	// Validate gender if provided
	if in.Gender != nil {
		switch *in.Gender {
		case domain.GenderMale, domain.GenderFemale, domain.GenderOther:
		default:
			return nil, fmt.Errorf("%w: invalid gender value", apierr.ErrValidation)
		}
	}

	// Strip non-digits from phone and validate length
	var cleanPhone *string
	if in.Phone != nil && *in.Phone != "" {
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' { return r }
			return -1
		}, *in.Phone)
		if len(digits) < 10 {
			return nil, fmt.Errorf("%w: phone must be at least 10 digits", apierr.ErrValidation)
		}
		cleanPhone = &digits
	}

	if err := s.users.UpdateEducationLevel(ctx, userID, in.EducationLevel); err != nil {
		return nil, fmt.Errorf("update education level: %w", err)
	}
	if err := s.users.UpdateProfile(ctx, userID, in.Gender, cleanPhone, in.AvatarURL); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	user.EducationLevel = in.EducationLevel
	user.Gender = in.Gender
	user.Phone = cleanPhone
	user.AvatarURL = in.AvatarURL
	return user, nil
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("get me: %w", err)
	}
	return user, nil
}

func (s *Service) issueTokenPair(ctx context.Context, user *domain.User) (TokenPair, error) {
	accessToken, err := pkgjwt.GenerateAccessToken(s.cfg.JWTSecret, s.cfg.JWTAccessTTL, user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate access token: %w", err)
	}

	rawRefresh, err := pkgjwt.GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now().UTC()
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: now.Add(s.cfg.JWTRefreshTTL),
		CreatedAt: now,
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return TokenPair{}, fmt.Errorf("store refresh token: %w", err)
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func validateRegisterInput(in RegisterInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)

	if utf8.RuneCountInString(in.Name) < 2 {
		return fmt.Errorf("%w: name must be at least 2 characters", apierr.ErrValidation)
	}
	if !strings.Contains(in.Email, "@") {
		return fmt.Errorf("%w: invalid email", apierr.ErrValidation)
	}
	if len(in.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", apierr.ErrValidation)
	}
	if in.Role != domain.RoleStudent && in.Role != domain.RoleContributor {
		return fmt.Errorf("%w: role must be student or contributor", apierr.ErrValidation)
	}
	return nil
}
