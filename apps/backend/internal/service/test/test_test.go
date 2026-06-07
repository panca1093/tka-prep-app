package test

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

type fakeTestRepo struct {
	tests     map[uuid.UUID]*domain.Test
	statuses  map[uuid.UUID]domain.TestStatus
	pubTimes  map[uuid.UUID]*time.Time
	deleted   []uuid.UUID
	qcount    map[uuid.UUID]int
}

func newFakeTestRepo(tests ...*domain.Test) *fakeTestRepo {
	r := &fakeTestRepo{
		tests:    make(map[uuid.UUID]*domain.Test),
		statuses: make(map[uuid.UUID]domain.TestStatus),
		pubTimes: make(map[uuid.UUID]*time.Time),
		qcount:   make(map[uuid.UUID]int),
	}
	for _, t := range tests {
		r.tests[t.ID] = t
		r.statuses[t.ID] = t.Status
		r.qcount[t.ID] = 3 // default non-zero
	}
	return r
}

func (r *fakeTestRepo) Create(_ context.Context, t *domain.Test) error {
	r.tests[t.ID] = t
	r.statuses[t.ID] = t.Status
	return nil
}

func (r *fakeTestRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Test, error) {
	t, ok := r.tests[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *fakeTestRepo) List(_ context.Context, _ repository.TestFilter) ([]*domain.Test, int, error) {
	return nil, 0, nil
}

func (r *fakeTestRepo) Update(_ context.Context, t *domain.Test) error {
	r.tests[t.ID] = t
	r.statuses[t.ID] = t.Status
	return nil
}

func (r *fakeTestRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *fakeTestRepo) SetStatus(_ context.Context, id uuid.UUID, status domain.TestStatus, publishedAt *time.Time) error {
	r.statuses[id] = status
	if publishedAt != nil {
		r.pubTimes[id] = publishedAt
	}
	return nil
}

func (r *fakeTestRepo) SetQuestions(_ context.Context, _ uuid.UUID, _ []domain.TestQuestion) error {
	return nil
}

func (r *fakeTestRepo) UpdateScoringConfig(_ context.Context, _ *domain.ScoringConfig) error {
	return nil
}

func (r *fakeTestRepo) QuestionCount(_ context.Context, id uuid.UUID) (int, error) {
	return r.qcount[id], nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newUUID() uuid.UUID { return uuid.New() }

func draftTest(ownerID uuid.UUID) *domain.Test {
	return &domain.Test{
		ID:              newUUID(),
		ContributorID:   ownerID,
		Title:           "Draft Test",
		CategoryID:      newUUID(),
		DurationMinutes: 60,
		Difficulty:      domain.DifficultyMedium,
		Status:          domain.TestStatusDraft,
		ScoringConfig: &domain.ScoringConfig{
			ID:            newUUID(),
			CorrectPoints: 4,
			WrongPoints:   0,
			BlankPoints:   0,
		},
	}
}

func publishedTest(ownerID uuid.UUID) *domain.Test {
	t := draftTest(ownerID)
	t.Status = domain.TestStatusPublished
	return t
}

func basicCreateInput(ownerID uuid.UUID) CreateInput {
	return CreateInput{
		ContributorID:   ownerID,
		Title:           "New Test",
		CategoryID:      newUUID(),
		DurationMinutes: 90,
		Difficulty:      domain.DifficultyHard,
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("C-01: creates a draft test with scoring config", func(t *testing.T) {
		repo := newFakeTestRepo()
		svc := NewService(repo)
		in := basicCreateInput(newUUID())
		in.ScoringConfig = &ScoringConfigInput{
			CorrectPoints: 4,
			WrongPoints:   -1,
			BlankPoints:   0,
		}

		got, err := svc.Create(ctx, in)

		require.NoError(t, err)
		require.Equal(t, "New Test", got.Title)
		require.Equal(t, domain.TestStatusDraft, got.Status)
		require.NotNil(t, got.ScoringConfig)
		require.Equal(t, float64(4), got.ScoringConfig.CorrectPoints)
		require.Equal(t, float64(-1), got.ScoringConfig.WrongPoints)
	})

	t.Run("C-02: defaults scoring to +4/0/0 when not provided", func(t *testing.T) {
		repo := newFakeTestRepo()
		svc := NewService(repo)
		in := basicCreateInput(newUUID())

		got, err := svc.Create(ctx, in)

		require.NoError(t, err)
		require.Equal(t, float64(4), got.ScoringConfig.CorrectPoints)
		require.Equal(t, float64(0), got.ScoringConfig.WrongPoints)
	})

	t.Run("C-03: rejects empty title", func(t *testing.T) {
		svc := NewService(newFakeTestRepo())
		in := basicCreateInput(newUUID())
		in.Title = "   "

		_, err := svc.Create(ctx, in)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("C-04: rejects non-positive duration", func(t *testing.T) {
		svc := NewService(newFakeTestRepo())
		in := basicCreateInput(newUUID())
		in.DurationMinutes = 0

		_, err := svc.Create(ctx, in)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})
}

// ─── Update ──────────────────────────────────────────────────────────────────

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("U-01: owner updates draft title", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		repo := newFakeTestRepo(dt)
		svc := NewService(repo)
		in := UpdateInput{Title: ptr("Updated Title")}

		got, err := svc.Update(ctx, dt.ID, owner, domain.RoleContributor, in)

		require.NoError(t, err)
		require.Equal(t, "Updated Title", got.Title)
	})

	t.Run("U-02: non-owner contributor cannot update", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		other := newUUID()
		repo := newFakeTestRepo(dt)
		svc := NewService(repo)

		_, err := svc.Update(ctx, dt.ID, other, domain.RoleContributor, UpdateInput{Title: ptr("Hack")})

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("U-03: admin can update any test", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		admin := newUUID()
		repo := newFakeTestRepo(dt)
		svc := NewService(repo)

		got, err := svc.Update(ctx, dt.ID, admin, domain.RoleAdmin, UpdateInput{Title: ptr("Admin Edit")})

		require.NoError(t, err)
		require.Equal(t, "Admin Edit", got.Title)
	})

	t.Run("U-04: cannot edit published test", func(t *testing.T) {
		owner := newUUID()
		pt := publishedTest(owner)
		repo := newFakeTestRepo(pt)
		svc := NewService(repo)

		_, err := svc.Update(ctx, pt.ID, owner, domain.RoleContributor, UpdateInput{Title: ptr("Try")})

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})
}

// ─── Publish ─────────────────────────────────────────────────────────────────

func TestPublish(t *testing.T) {
	ctx := context.Background()

	t.Run("P-01: owner publishes draft with questions", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		repo := newFakeTestRepo(dt)
		repo.qcount[dt.ID] = 5
		svc := NewService(repo)

		got, err := svc.Publish(ctx, dt.ID, owner, domain.RoleContributor)

		require.NoError(t, err)
		require.Equal(t, domain.TestStatusPublished, got.Status)
	})

	t.Run("P-02: cannot publish with zero questions", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		repo := newFakeTestRepo(dt)
		repo.qcount[dt.ID] = 0
		svc := NewService(repo)

		_, err := svc.Publish(ctx, dt.ID, owner, domain.RoleContributor)

		require.Error(t, err)
		require.True(t, errors.Is(err, apierr.ErrValidation))
	})

	t.Run("P-03: non-owner cannot publish", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		svc := NewService(newFakeTestRepo(dt))

		_, err := svc.Publish(ctx, dt.ID, newUUID(), domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("P-04: admin can publish any draft", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		repo := newFakeTestRepo(dt)
		repo.qcount[dt.ID] = 3
		svc := NewService(repo)

		got, err := svc.Publish(ctx, dt.ID, newUUID(), domain.RoleAdmin)

		require.NoError(t, err)
		require.Equal(t, domain.TestStatusPublished, got.Status)
	})

	t.Run("P-05: already published returns conflict", func(t *testing.T) {
		owner := newUUID()
		pt := publishedTest(owner)
		svc := NewService(newFakeTestRepo(pt))

		_, err := svc.Publish(ctx, pt.ID, owner, domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})
}

// ─── Unpublish ───────────────────────────────────────────────────────────────

func TestUnpublish(t *testing.T) {
	ctx := context.Background()

	t.Run("UP-01: owner unpublishes own test", func(t *testing.T) {
		owner := newUUID()
		pt := publishedTest(owner)
		repo := newFakeTestRepo(pt)
		svc := NewService(repo)

		got, err := svc.Unpublish(ctx, pt.ID, owner, domain.RoleContributor)

		require.NoError(t, err)
		require.Equal(t, domain.TestStatusDraft, got.Status)
	})

	t.Run("UP-02: non-owner contributor cannot unpublish", func(t *testing.T) {
		owner := newUUID()
		pt := publishedTest(owner)
		svc := NewService(newFakeTestRepo(pt))

		_, err := svc.Unpublish(ctx, pt.ID, newUUID(), domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("UP-03: admin can unpublish any published test", func(t *testing.T) {
		owner := newUUID()
		pt := publishedTest(owner)
		repo := newFakeTestRepo(pt)
		svc := NewService(repo)

		got, err := svc.Unpublish(ctx, pt.ID, newUUID(), domain.RoleAdmin)

		require.NoError(t, err)
		require.Equal(t, domain.TestStatusDraft, got.Status)
	})

	t.Run("UP-04: already draft returns conflict", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		svc := NewService(newFakeTestRepo(dt))

		_, err := svc.Unpublish(ctx, dt.ID, owner, domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("D-01: owner deletes own draft", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		repo := newFakeTestRepo(dt)
		svc := NewService(repo)

		err := svc.Delete(ctx, dt.ID, owner, domain.RoleContributor)

		require.NoError(t, err)
		require.Contains(t, repo.deleted, dt.ID)
	})

	t.Run("D-02: cannot delete published test", func(t *testing.T) {
		owner := newUUID()
		pt := publishedTest(owner)
		svc := NewService(newFakeTestRepo(pt))

		err := svc.Delete(ctx, pt.ID, owner, domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("D-03: non-owner cannot delete", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		svc := NewService(newFakeTestRepo(dt))

		err := svc.Delete(ctx, dt.ID, newUUID(), domain.RoleContributor)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("D-04: admin can delete any draft", func(t *testing.T) {
		owner := newUUID()
		dt := draftTest(owner)
		repo := newFakeTestRepo(dt)
		svc := NewService(repo)

		err := svc.Delete(ctx, dt.ID, newUUID(), domain.RoleAdmin)

		require.NoError(t, err)
		require.Contains(t, repo.deleted, dt.ID)
	})
}

func ptr[T any](v T) *T { return &v }
