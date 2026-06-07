package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

// ─── fakes ───────────────────────────────────────────────────────────────────

type fakeUserRepo struct {
	users       map[uuid.UUID]*domain.User
	updateCalls []statusUpdateCall
}

type statusUpdateCall struct {
	UserID uuid.UUID
	Status domain.Status
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if u.Email == "exists@test.dev" {
		return apierr.ErrConflict
	}
	r.users[u.ID] = u
	return nil
}
func (r *fakeUserRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, apierr.ErrNotFound
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
	if _, ok := r.users[id]; !ok {
		return apierr.ErrNotFound
	}
	r.updateCalls = append(r.updateCalls, statusUpdateCall{UserID: id, Status: status})
	return nil
}
func (r *fakeUserRepo) UpdateEducationLevel(_ context.Context, _ uuid.UUID, _ *domain.EducationLevel) error {
	return nil
}
func (r *fakeUserRepo) UpdateProfile(_ context.Context, _ uuid.UUID, _ *domain.Gender, _ *string, _ *string) error {
	return nil
}

type fakeAdminRepo struct{}

func (r *fakeAdminRepo) GetStats(_ context.Context) (*domain.PlatformStats, error) {
	return &domain.PlatformStats{}, nil
}
func (r *fakeAdminRepo) ListTestsWithAttempts(_ context.Context, _, _ int, _ *domain.EducationLevel, _ *uuid.UUID) ([]*domain.TestWithAttempts, int, error) {
	return nil, 0, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newUUID() uuid.UUID { return uuid.New() }

func pendingContributor() *domain.User {
	return &domain.User{
		ID:    newUUID(),
		Name:  "Pending Guy",
		Email: "pending@test.dev",
		Role:  domain.RoleContributor,
		Status: domain.StatusPending,
	}
}

func activeContributor() *domain.User {
	return &domain.User{
		ID:    newUUID(),
		Name:  "Active Guy",
		Email: "active@test.dev",
		Role:  domain.RoleContributor,
		Status: domain.StatusActive,
	}
}

func studentUser() *domain.User {
	return &domain.User{
		ID:    newUUID(),
		Name:  "Student",
		Email: "student@test.dev",
		Role:  domain.RoleStudent,
		Status: domain.StatusActive,
	}
}

func svc(repo *fakeUserRepo) *Service {
	return New(repo, &fakeAdminRepo{})
}

// ─── ApproveContributor ──────────────────────────────────────────────────────

func TestApproveContributor(t *testing.T) {
	ctx := context.Background()

	t.Run("A-01: approves pending contributor", func(t *testing.T) {
		c := pendingContributor()
		repo := newFakeUserRepo()
		repo.users[c.ID] = c
		s := svc(repo)

		err := s.ApproveContributor(ctx, c.ID)

		require.NoError(t, err)
		require.Len(t, repo.updateCalls, 1)
		require.Equal(t, domain.StatusActive, repo.updateCalls[0].Status)
	})

	t.Run("A-02: rejects non-contributor", func(t *testing.T) {
		u := studentUser()
		repo := newFakeUserRepo()
		repo.users[u.ID] = u
		s := svc(repo)

		err := s.ApproveContributor(ctx, u.ID)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("A-03: rejects already-active contributor", func(t *testing.T) {
		c := activeContributor()
		repo := newFakeUserRepo()
		repo.users[c.ID] = c
		s := svc(repo)

		err := s.ApproveContributor(ctx, c.ID)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("A-04: returns error for unknown user", func(t *testing.T) {
		s := svc(newFakeUserRepo())

		err := s.ApproveContributor(ctx, newUUID())
		require.Error(t, err)
	})
}

// ─── RejectContributor ───────────────────────────────────────────────────────

func TestRejectContributor(t *testing.T) {
	ctx := context.Background()

	t.Run("RJ-01: rejects pending contributor (suspends)", func(t *testing.T) {
		c := pendingContributor()
		repo := newFakeUserRepo()
		repo.users[c.ID] = c
		s := svc(repo)

		err := s.RejectContributor(ctx, c.ID)

		require.NoError(t, err)
		require.Len(t, repo.updateCalls, 1)
		require.Equal(t, domain.StatusSuspended, repo.updateCalls[0].Status)
	})

	t.Run("RJ-02: rejects non-contributor", func(t *testing.T) {
		u := studentUser()
		repo := newFakeUserRepo()
		repo.users[u.ID] = u
		s := svc(repo)

		err := s.RejectContributor(ctx, u.ID)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}

// ─── UpdateUserStatus ────────────────────────────────────────────────────────

func TestUpdateUserStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("US-01: admin suspends an active user", func(t *testing.T) {
		u := studentUser()
		repo := newFakeUserRepo()
		repo.users[u.ID] = u
		s := svc(repo)

		err := s.UpdateUserStatus(ctx, u.ID, domain.StatusSuspended)

		require.NoError(t, err)
		require.Len(t, repo.updateCalls, 1)
		require.Equal(t, domain.StatusSuspended, repo.updateCalls[0].Status)
	})

	t.Run("US-02: returns not found for unknown user", func(t *testing.T) {
		s := svc(newFakeUserRepo())

		err := s.UpdateUserStatus(ctx, newUUID(), domain.StatusActive)
		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})

	t.Run("US-03: rejects invalid status", func(t *testing.T) {
		u := studentUser()
		repo := newFakeUserRepo()
		repo.users[u.ID] = u
		s := svc(repo)

		err := s.UpdateUserStatus(ctx, u.ID, "bogus")
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}

// ─── CreateContributor ───────────────────────────────────────────────────────

func TestCreateContributor(t *testing.T) {
	ctx := context.Background()

	t.Run("CC-01: admin creates contributor directly active", func(t *testing.T) {
		repo := newFakeUserRepo()
		s := svc(repo)

		user, err := s.CreateContributor(ctx, "New Contrib", "newcontrib@test.dev", "password123")

		require.NoError(t, err)
		require.Equal(t, "New Contrib", user.Name)
		require.Equal(t, domain.RoleContributor, user.Role)
		require.Equal(t, domain.StatusActive, user.Status)
	})

	t.Run("CC-02: rejects short password", func(t *testing.T) {
		s := svc(newFakeUserRepo())

		_, err := s.CreateContributor(ctx, "Name", "email@test.dev", "short")
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("CC-03: rejects invalid email", func(t *testing.T) {
		s := svc(newFakeUserRepo())

		_, err := s.CreateContributor(ctx, "Name", "bad-email", "password123")
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("CC-04: rejects empty name", func(t *testing.T) {
		s := svc(newFakeUserRepo())

		_, err := s.CreateContributor(ctx, "", "email@test.dev", "password123")
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}
