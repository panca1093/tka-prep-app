package topic

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

type fakeTopicRepo struct {
	topics  map[uuid.UUID]*domain.Topic
	created []*domain.Topic
	updated []*domain.Topic
	deleted []uuid.UUID
	used    map[uuid.UUID]bool
}

func newFakeTopicRepo(topics ...*domain.Topic) *fakeTopicRepo {
	r := &fakeTopicRepo{
		topics: make(map[uuid.UUID]*domain.Topic),
		used:   make(map[uuid.UUID]bool),
	}
	for _, t := range topics {
		cp := *t
		r.topics[cp.ID] = &cp
	}
	return r
}

func (r *fakeTopicRepo) List(_ context.Context) ([]*domain.Topic, error) { return nil, nil }
func (r *fakeTopicRepo) Create(_ context.Context, t *domain.Topic) error {
	if t.Name == "" {
		return apierr.ErrConflict
	}
	r.topics[t.ID] = t
	r.created = append(r.created, t)
	return nil
}
func (r *fakeTopicRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Topic, error) {
	t, ok := r.topics[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	cp := *t
	return &cp, nil
}
func (r *fakeTopicRepo) Update(_ context.Context, t *domain.Topic) error {
	if _, ok := r.topics[t.ID]; !ok {
		return apierr.ErrNotFound
	}
	r.topics[t.ID] = t
	r.updated = append(r.updated, t)
	return nil
}
func (r *fakeTopicRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted = append(r.deleted, id)
	return nil
}
func (r *fakeTopicRepo) IsUsedByQuestions(_ context.Context, id uuid.UUID) (bool, error) {
	return r.used[id], nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newUUID() uuid.UUID { return uuid.New() }

func contributorOwnedTopic(ownerID uuid.UUID) *domain.Topic {
	return &domain.Topic{
		ID:          newUUID(),
		Name:        "Matematika",
		CreatedBy:   &ownerID,
	}
}

func seededTopic() *domain.Topic {
	return &domain.Topic{
		ID:        newUUID(),
		Name:      "Fisika",
		CreatedBy: nil, // admin-seeded, owned by no one
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("C-01: contributor creates topic with name", func(t *testing.T) {
		repo := newFakeTopicRepo()
		svc := NewService(repo)

		got, err := svc.Create(ctx, CreateInput{
			Name:      "Biologi",
			CreatedBy: newUUID(),
		})

		require.NoError(t, err)
		require.Equal(t, "Biologi", got.Name)
		require.NotNil(t, got.CreatedBy)
	})

	t.Run("C-02: rejects empty name", func(t *testing.T) {
		svc := NewService(newFakeTopicRepo())

		_, err := svc.Create(ctx, CreateInput{Name: "   ", CreatedBy: newUUID()})

		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}

// ─── Update ──────────────────────────────────────────────────────────────────

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("U-01: owner updates their topic name", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		repo := newFakeTopicRepo(topic)
		svc := NewService(repo)

		newName := "Matematika Lanjutan"
		got, err := svc.Update(ctx, topic.ID, owner, domain.RoleContributor, UpdateInput{Name: &newName})

		require.NoError(t, err)
		require.Equal(t, "Matematika Lanjutan", got.Name)
	})

	t.Run("U-02: non-owner contributor cannot update", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		svc := NewService(newFakeTopicRepo(topic))

		_, err := svc.Update(ctx, topic.ID, newUUID(), domain.RoleContributor, UpdateInput{Name: ptr("Hack")})

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("U-03: admin can update any topic", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		repo := newFakeTopicRepo(topic)
		svc := NewService(repo)

		got, err := svc.Update(ctx, topic.ID, newUUID(), domain.RoleAdmin, UpdateInput{Name: ptr("Admin Rename")})

		require.NoError(t, err)
		require.Equal(t, "Admin Rename", got.Name)
	})

	t.Run("U-04: contributor cannot update seeded topic", func(t *testing.T) {
		topic := seededTopic()
		svc := NewService(newFakeTopicRepo(topic))

		_, err := svc.Update(ctx, topic.ID, newUUID(), domain.RoleContributor, UpdateInput{Name: ptr("Try")})

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("U-05: rejects empty name update", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		svc := NewService(newFakeTopicRepo(topic))

		_, err := svc.Update(ctx, topic.ID, owner, domain.RoleContributor, UpdateInput{Name: ptr("  ")})

		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("D-01: owner deletes unused topic", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		repo := newFakeTopicRepo(topic)
		svc := NewService(repo)

		err := svc.Delete(ctx, topic.ID, owner, domain.RoleContributor)

		require.NoError(t, err)
		require.Contains(t, repo.deleted, topic.ID)
	})

	t.Run("D-02: cannot delete topic used by questions", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		repo := newFakeTopicRepo(topic)
		repo.used[topic.ID] = true
		svc := NewService(repo)

		err := svc.Delete(ctx, topic.ID, owner, domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("D-03: non-owner contributor cannot delete", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		svc := NewService(newFakeTopicRepo(topic))

		err := svc.Delete(ctx, topic.ID, newUUID(), domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("D-04: admin can delete any topic", func(t *testing.T) {
		owner := newUUID()
		topic := contributorOwnedTopic(owner)
		repo := newFakeTopicRepo(topic)
		svc := NewService(repo)

		err := svc.Delete(ctx, topic.ID, newUUID(), domain.RoleAdmin)

		require.NoError(t, err)
		require.Contains(t, repo.deleted, topic.ID)
	})

	t.Run("D-05: cannot delete seeded topic as non-admin", func(t *testing.T) {
		topic := seededTopic()
		svc := NewService(newFakeTopicRepo(topic))

		err := svc.Delete(ctx, topic.ID, newUUID(), domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})
}

func ptr[T any](v T) *T { return &v }
