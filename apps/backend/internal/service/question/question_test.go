package question

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

// ─── fakes ─────────────────────────────────────────────────────────────────────

type fakeQuestionRepo struct {
	questions map[uuid.UUID]*domain.Question
	// listFilters captures the last List call's filter for assertions.
	listFilters []repository.QuestionFilter
}

func newFakeQuestionRepo(qs ...*domain.Question) *fakeQuestionRepo {
	r := &fakeQuestionRepo{questions: make(map[uuid.UUID]*domain.Question)}
	for _, q := range qs {
		cp := *q
		r.questions[cp.ID] = &cp
	}
	return r
}

func (r *fakeQuestionRepo) List(_ context.Context, f repository.QuestionFilter) ([]*domain.Question, int, error) {
	r.listFilters = append(r.listFilters, f)

	// Simulate the SQL filtering that the real repo would do.
	var out []*domain.Question
	for _, q := range r.questions {
		if f.CallerRole != domain.RoleAdmin && q.ContributorID != f.CallerID {
			continue
		}
		if f.TopicID != nil && q.TopicID != *f.TopicID {
			continue
		}
		if f.Difficulty != nil && q.Difficulty != *f.Difficulty {
			continue
		}
		if f.QuestionType != nil && q.Type != *f.QuestionType {
			continue
		}
		if f.EducationLevel != nil {
			if q.EducationLevel == nil || *q.EducationLevel != *f.EducationLevel {
				continue
			}
		}
		if f.Search != "" {
			// Simplistic: match against text only.
			if !containsFold(q.Text, f.Search) {
				continue
			}
		}
		out = append(out, q)
	}
	return out, len(out), nil
}

func (r *fakeQuestionRepo) Create(_ context.Context, _ *domain.Question) error { return nil }
func (r *fakeQuestionRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Question, error) {
	q, ok := r.questions[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	cp := *q
	cp.Options = make([]domain.QuestionOption, len(q.Options))
	copy(cp.Options, q.Options)
	cp.Statements = make([]domain.QuestionStatement, len(q.Statements))
	copy(cp.Statements, q.Statements)
	return &cp, nil
}
func (r *fakeQuestionRepo) Update(_ context.Context, _ *domain.Question) error { return nil }
func (r *fakeQuestionRepo) Delete(_ context.Context, _ uuid.UUID) error        { return nil }
func (r *fakeQuestionRepo) IsUsedInPublishedTest(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (r *fakeQuestionRepo) ListUsageStats(_ context.Context, _ uuid.UUID, _ int) ([]domain.QuestionUsageEntry, error) {
	return nil, nil
}

// ─── helpers ───────────────────────────────────────────────────────────────────

func makeQuestion(id, contributorID uuid.UUID) *domain.Question {
	return &domain.Question{
		ID:            id,
		ContributorID: contributorID,
		TopicID:       uuid.New(),
		Type:          domain.QuestionTypeMCQ,
		Text:          "Apa itu?",
		Difficulty:    domain.DifficultyEasy,
		Options: []domain.QuestionOption{
			{ID: uuid.New(), Label: "A", Text: "Satu", IsCorrect: true},
			{ID: uuid.New(), Label: "B", Text: "Dua", IsCorrect: false},
			{ID: uuid.New(), Label: "C", Text: "Tiga", IsCorrect: false},
			{ID: uuid.New(), Label: "D", Text: "Empat", IsCorrect: false},
		},
	}
}

func containsFold(s, substr string) bool {
	return len(s) >= len(substr) && searchFold(s, substr)
}

func searchFold(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// ─── TestGet ───────────────────────────────────────────────────────────────────

func TestGet(t *testing.T) {
	ctx := context.Background()

	t.Run("contributor sees own question", func(t *testing.T) {
		contributorID := uuid.New()
		q := makeQuestion(uuid.New(), contributorID)
		svc := NewService(newFakeQuestionRepo(q), nil)

		got, err := svc.Get(ctx, q.ID, contributorID, domain.RoleContributor)
		require.NoError(t, err)
		require.Equal(t, q.ID, got.ID)
	})

	t.Run("contributor gets ErrNotFound for non-owned question", func(t *testing.T) {
		ownerID := uuid.New()
		callerID := uuid.New()
		q := makeQuestion(uuid.New(), ownerID)
		svc := NewService(newFakeQuestionRepo(q), nil)

		_, err := svc.Get(ctx, q.ID, callerID, domain.RoleContributor)
		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})

	t.Run("admin sees any question", func(t *testing.T) {
		ownerID := uuid.New()
		adminID := uuid.New()
		q := makeQuestion(uuid.New(), ownerID)
		svc := NewService(newFakeQuestionRepo(q), nil)

		got, err := svc.Get(ctx, q.ID, adminID, domain.RoleAdmin)
		require.NoError(t, err)
		require.Equal(t, q.ID, got.ID)
	})

	t.Run("student sees any question (test-taking bypass)", func(t *testing.T) {
		ownerID := uuid.New()
		studentID := uuid.New()
		q := makeQuestion(uuid.New(), ownerID)
		svc := NewService(newFakeQuestionRepo(q), nil)

		got, err := svc.Get(ctx, q.ID, studentID, domain.RoleStudent)
		require.NoError(t, err)
		require.Equal(t, q.ID, got.ID)
	})

	t.Run("ErrNotFound when question does not exist", func(t *testing.T) {
		svc := NewService(newFakeQuestionRepo(), nil)
		_, err := svc.Get(ctx, uuid.New(), uuid.New(), domain.RoleAdmin)
		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})
}

// ─── TestList_VisibilityFilter ─────────────────────────────────────────────────

func TestList_VisibilityFilter(t *testing.T) {
	ctx := context.Background()

	contributorA := uuid.New()
	contributorB := uuid.New()
	adminID := uuid.New()

	qA1 := makeQuestion(uuid.New(), contributorA)
	qA2 := makeQuestion(uuid.New(), contributorA)
	qB1 := makeQuestion(uuid.New(), contributorB)

	repo := newFakeQuestionRepo(qA1, qA2, qB1)

	t.Run("contributor sees only own questions", func(t *testing.T) {
		svc := NewService(repo, nil)
		// Reset listFilters for clean assertion.
		repo.listFilters = nil

		qs, total, err := svc.List(ctx, ListFilter{
			CallerID:   contributorA,
			CallerRole: domain.RoleContributor,
			Page:       1,
			Limit:      20,
		})
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, qs, 2)
		for _, q := range qs {
			require.Equal(t, contributorA, q.ContributorID)
		}
	})

	t.Run("admin sees all questions", func(t *testing.T) {
		svc := NewService(repo, nil)
		repo.listFilters = nil

		qs, total, err := svc.List(ctx, ListFilter{
			CallerID:   adminID,
			CallerRole: domain.RoleAdmin,
			Page:       1,
			Limit:      20,
		})
		require.NoError(t, err)
		require.Equal(t, 3, total)
		require.Len(t, qs, 3)
	})
}
