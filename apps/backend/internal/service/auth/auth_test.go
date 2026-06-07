package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

// ─── fakes ───────────────────────────────────────────────────────────────────

type fakeUserRepo struct {
	users       map[uuid.UUID]*domain.User
	byEmail     map[string]*domain.User
	created     []*domain.User
	updateCalls []statusUpdateCall
}

type statusUpdateCall struct {
	UserID uuid.UUID
	Status domain.Status
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if _, ok := r.byEmail[u.Email]; ok {
		return apierr.ErrConflict
	}
	r.users[u.ID] = u
	r.byEmail[u.Email] = u
	r.created = append(r.created, u)
	return nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) List(_ context.Context, _ repository.UserAdminFilter) ([]*domain.User, int, error) {
	return nil, 0, nil
}

func (r *fakeUserRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.Status) error {
	u, ok := r.users[id]
	if !ok {
		return apierr.ErrNotFound
	}
	u.Status = status
	r.updateCalls = append(r.updateCalls, statusUpdateCall{UserID: id, Status: status})
	return nil
}

func (r *fakeUserRepo) UpdateEducationLevel(_ context.Context, _ uuid.UUID, _ *domain.EducationLevel) error {
	return nil
}

func (r *fakeUserRepo) UpdateProfile(_ context.Context, _ uuid.UUID, _ *domain.Gender, _ *string, _ *string) error {
	return nil
}

type fakeTokenRepo struct {
	tokens  map[string]*domain.RefreshToken
	deleted []string
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{tokens: make(map[string]*domain.RefreshToken)}
}

func (r *fakeTokenRepo) Create(_ context.Context, t *domain.RefreshToken) error {
	r.tokens[t.TokenHash] = t
	return nil
}

func (r *fakeTokenRepo) FindByTokenHash(_ context.Context, hash string) (*domain.RefreshToken, error) {
	t, ok := r.tokens[hash]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	return t, nil
}

func (r *fakeTokenRepo) DeleteByTokenHash(_ context.Context, hash string) error {
	r.deleted = append(r.deleted, hash)
	return nil
}

func (r *fakeTokenRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func cfg() Config {
	return Config{
		JWTSecret:     "test-secret-for-testing-purposes-32b",
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
	}
}

func validStudent() RegisterInput {
	return RegisterInput{
		Name:     "Test Student",
		Email:    "student@test.dev",
		Password: "password123",
		Role:     domain.RoleStudent,
	}
}

func validContributor() RegisterInput {
	return RegisterInput{
		Name:     "Test Contributor",
		Email:    "contrib@test.dev",
		Password: "password123",
		Role:     domain.RoleContributor,
	}
}

func ptr[T any](v T) *T { return &v }

// ─── Register ────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("R-01: student self-registers and receives token pair", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())
		in := validStudent()

		user, pair, err := svc.Register(ctx, in)

		require.NoError(t, err)
		require.Equal(t, "Test Student", user.Name)
		require.Equal(t, "student@test.dev", user.Email)
		require.Equal(t, domain.RoleStudent, user.Role)
		require.Equal(t, domain.StatusActive, user.Status)
		require.NotEmpty(t, pair.AccessToken)
		require.NotEmpty(t, pair.RefreshToken)
	})

	t.Run("R-02: contributor self-registers with pending status and no tokens", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())
		in := validContributor()

		user, pair, err := svc.Register(ctx, in)

		require.NoError(t, err)
		require.Equal(t, domain.RoleContributor, user.Role)
		require.Equal(t, domain.StatusPending, user.Status)
		require.Empty(t, pair.AccessToken, "pending contributor must not receive tokens")
	})

	t.Run("R-03: rejects duplicate email", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())

		_, _, err := svc.Register(ctx, validStudent())
		require.NoError(t, err)

		_, _, err = svc.Register(ctx, validStudent())
		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("R-04: rejects name shorter than 2 characters", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())
		in := validStudent()
		in.Name = "A"

		_, _, err := svc.Register(ctx, in)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("R-05: rejects invalid email", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())
		in := validStudent()
		in.Email = "not-an-email"

		_, _, err := svc.Register(ctx, in)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("R-06: rejects password shorter than 8 characters", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())
		in := validStudent()
		in.Password = "short"

		_, _, err := svc.Register(ctx, in)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("R-07: rejects invalid role", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())
		in := validStudent()
		in.Role = domain.RoleAdmin

		_, _, err := svc.Register(ctx, in)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("R-08: student with education level set gets it persisted", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())
		in := validStudent()
		el := domain.EducationLevelSMP
		in.EducationLevel = &el

		user, _, err := svc.Register(ctx, in)

		require.NoError(t, err)
		require.NotNil(t, user.EducationLevel)
		require.Equal(t, domain.EducationLevelSMP, *user.EducationLevel)
	})
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("L-01: active student logs in with correct password", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())
		_, _, err := svc.Register(ctx, validStudent())
		require.NoError(t, err)

		user, pair, err := svc.Login(ctx, "student@test.dev", "password123")

		require.NoError(t, err)
		require.Equal(t, "student@test.dev", user.Email)
		require.NotEmpty(t, pair.AccessToken)
	})

	t.Run("L-02: wrong password returns ErrUnauthorized", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())
		_, _, err := svc.Register(ctx, validStudent())
		require.NoError(t, err)

		_, _, err = svc.Login(ctx, "student@test.dev", "wrongpassword")
		require.True(t, errors.Is(err, apierr.ErrUnauthorized))
	})

	t.Run("L-03: unknown email returns ErrUnauthorized", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())

		_, _, err := svc.Login(ctx, "noone@test.dev", "password123")
		require.True(t, errors.Is(err, apierr.ErrUnauthorized))
	})

	t.Run("L-04: pending contributor cannot log in", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())
		_, _, err := svc.Register(ctx, validContributor())
		require.NoError(t, err)

		_, _, err = svc.Login(ctx, "contrib@test.dev", "password123")
		require.True(t, errors.Is(err, apierr.ErrPending))
	})

	t.Run("L-05: suspended user cannot log in", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())
		user, _, err := svc.Register(ctx, validStudent())
		require.NoError(t, err)
		_ = repo.UpdateStatus(ctx, user.ID, domain.StatusSuspended)

		_, _, err = svc.Login(ctx, "student@test.dev", "password123")
		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})
}

// ─── Logout ──────────────────────────────────────────────────────────────────

func TestLogout(t *testing.T) {
	ctx := context.Background()

	t.Run("LO-01: logout deletes the refresh token", func(t *testing.T) {
		repo := newFakeUserRepo()
		tokens := newFakeTokenRepo()
		svc := NewService(cfg(), repo, tokens)
		_, pair, err := svc.Register(ctx, validStudent())
		require.NoError(t, err)

		err = svc.Logout(ctx, pair.RefreshToken)
		require.NoError(t, err)
		require.Len(t, tokens.deleted, 1)
	})
}

// ─── Me ──────────────────────────────────────────────────────────────────────

func TestMe(t *testing.T) {
	ctx := context.Background()

	t.Run("M-01: returns user by ID", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewService(cfg(), repo, newFakeTokenRepo())
		created, _, err := svc.Register(ctx, validStudent())
		require.NoError(t, err)

		user, err := svc.Me(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.Email, user.Email)
	})

	t.Run("M-02: returns ErrNotFound for unknown ID", func(t *testing.T) {
		svc := NewService(cfg(), newFakeUserRepo(), newFakeTokenRepo())

		_, err := svc.Me(ctx, uuid.New())
		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})
}
