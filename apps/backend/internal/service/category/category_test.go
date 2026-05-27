package category

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
)

// ─── fakes ───────────────────────────────────────────────────────────────────

type fakeCategoryRepo struct {
	categories map[uuid.UUID]*domain.Category
	updated    []updatedCall
}

type updatedCall struct {
	callerID    uuid.UUID
	id          uuid.UUID
	name        string
	description *string
}

func newFakeRepo(cats ...*domain.Category) *fakeCategoryRepo {
	r := &fakeCategoryRepo{categories: make(map[uuid.UUID]*domain.Category)}
	for _, c := range cats {
		cp := *c
		r.categories[cp.ID] = &cp
	}
	return r
}

func (r *fakeCategoryRepo) List(_ context.Context) ([]*domain.Category, error) { return nil, nil }
func (r *fakeCategoryRepo) Create(_ context.Context, _ string) (*domain.Category, error) {
	return nil, nil
}
func (r *fakeCategoryRepo) CreateOwned(_ context.Context, _ uuid.UUID, _ string, _ *string) (*domain.Category, error) {
	return &domain.Category{ID: uuid.New()}, nil
}
func (r *fakeCategoryRepo) Update(_ context.Context, _ uuid.UUID, _ string) (*domain.Category, error) {
	return nil, nil
}
func (r *fakeCategoryRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (r *fakeCategoryRepo) TestCount(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (r *fakeCategoryRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Category, error) {
	c, ok := r.categories[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	cp := *c
	return &cp, nil
}
func (r *fakeCategoryRepo) UpdateOwned(_ context.Context, callerID, id uuid.UUID, name string, description *string) (*domain.Category, error) {
	c, ok := r.categories[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	if c.CreatedBy == nil || *c.CreatedBy != callerID {
		return nil, apierr.ErrForbidden
	}
	r.updated = append(r.updated, updatedCall{callerID: callerID, id: id, name: name, description: description})
	cp := *c
	cp.Name = name
	cp.Description = description
	return &cp, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func ptr[T any](v T) *T { return &v }

func ownedCategory(ownerID uuid.UUID) *domain.Category {
	return &domain.Category{
		ID:        uuid.New(),
		Name:      "Kategori Lama",
		CreatedBy: &ownerID,
	}
}

func seededCategory() *domain.Category {
	return &domain.Category{
		ID:        uuid.New(),
		Name:      "tka_saintek",
		CreatedBy: nil, // admin-owned / seeded
	}
}

// ─── TestUpdateOwned ─────────────────────────────────────────────────────────

func TestUpdateOwned(t *testing.T) {
	ctx := context.Background()

	t.Run("C-01: contributor updates their own category", func(t *testing.T) {
		ownerID := uuid.New()
		cat := ownedCategory(ownerID)
		svc := New(newFakeRepo(cat))

		got, err := svc.UpdateOwned(ctx, ownerID, cat.ID, "Kategori Baru", ptr("Deskripsi baru"))

		require.NoError(t, err)
		require.Equal(t, "Kategori Baru", got.Name)
		require.NotNil(t, got.Description)
		require.Equal(t, "Deskripsi baru", *got.Description)
	})

	t.Run("C-02: returns ErrForbidden when category is owned by someone else", func(t *testing.T) {
		ownerID, otherID := uuid.New(), uuid.New()
		cat := ownedCategory(ownerID)
		svc := New(newFakeRepo(cat))

		_, err := svc.UpdateOwned(ctx, otherID, cat.ID, "Hack", nil)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("C-03: returns ErrForbidden for seeded category (created_by = NULL)", func(t *testing.T) {
		cat := seededCategory()
		callerID := uuid.New()
		svc := New(newFakeRepo(cat))

		_, err := svc.UpdateOwned(ctx, callerID, cat.ID, "Renamed", nil)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("C-04: returns ErrNotFound when category does not exist", func(t *testing.T) {
		svc := New(newFakeRepo())

		_, err := svc.UpdateOwned(ctx, uuid.New(), uuid.New(), "X", nil)

		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})

	t.Run("C-05: returns ErrValidation when name is blank", func(t *testing.T) {
		ownerID := uuid.New()
		cat := ownedCategory(ownerID)
		svc := New(newFakeRepo(cat))

		_, err := svc.UpdateOwned(ctx, ownerID, cat.ID, "  ", nil)

		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}
