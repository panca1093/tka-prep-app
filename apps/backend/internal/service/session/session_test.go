package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
)

// ─── TestStart ───────────────────────────────────────────────────────────────

func TestStart(t *testing.T) {
	ctx := context.Background()

	t.Run("S-01: creates new session when no active session exists", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 60)), nil)

		got, err := svc.Start(ctx, studentID, testID)

		require.NoError(t, err)
		require.Equal(t, studentID, got.StudentID)
		require.Equal(t, testID, got.TestID)
		require.Equal(t, domain.SessionStatusInProgress, got.Status)
		require.Equal(t, 60*60, got.TimeRemainingSeconds)
		require.Len(t, sessRepo.created, 1)
	})

	t.Run("S-02: resumes existing in_progress session", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		existing := makeInProgressSession(uuid.New(), studentID, testID, 5*time.Minute)
		sessRepo.seed(existing)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 60)), nil)

		got, err := svc.Start(ctx, studentID, testID)

		require.NoError(t, err)
		require.Equal(t, existing.ID, got.ID, "must return the same session ID")
		require.Equal(t, domain.SessionStatusInProgress, got.Status)
		require.Empty(t, sessRepo.created, "must not create a new session")
		// started 5 min ago on a 60-min test → ~55 min left (allow 5s for test execution)
		require.InDelta(t, 55*60, got.TimeRemainingSeconds, 5)
	})

	t.Run("S-03: auto-expires overdue session and creates a fresh one", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		old := makeInProgressSession(uuid.New(), studentID, testID, 2*time.Minute)
		sessRepo.seed(old)
		// 1-minute test — session that started 2 min ago is overdue
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 1)), nil)

		got, err := svc.Start(ctx, studentID, testID)

		require.NoError(t, err)
		require.NotEqual(t, old.ID, got.ID, "must return a new session")
		require.Equal(t, domain.SessionStatusInProgress, got.Status)
		require.Len(t, sessRepo.statuses, 1)
		require.Equal(t, old.ID, sessRepo.statuses[0].id)
		require.Equal(t, domain.SessionStatusExpired, sessRepo.statuses[0].status)
		require.Len(t, sessRepo.created, 1)
	})

	t.Run("S-04: returns ErrNotFound when test does not exist", func(t *testing.T) {
		svc := NewService(newFakeSessionRepo(), &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.Start(ctx, uuid.New(), uuid.New())

		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})

	t.Run("S-05: returns ErrNotFound when test is a draft (not published)", func(t *testing.T) {
		testID := uuid.New()
		svc := NewService(newFakeSessionRepo(), &fakeResultRepo{}, newFakeTestRepo(makeDraftTest(testID)), nil)

		_, err := svc.Start(ctx, uuid.New(), testID)

		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})
}

// ─── TestGet ──────────────────────────────────────────────────────────────────

func TestGet(t *testing.T) {
	ctx := context.Background()

	t.Run("S-06: returns session with correct TimeRemainingSeconds", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), studentID, testID, 30*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 60)), nil)

		got, err := svc.Get(ctx, sess.ID, studentID)

		require.NoError(t, err)
		require.Equal(t, sess.ID, got.ID)
		require.Equal(t, domain.SessionStatusInProgress, got.Status)
		// started 30 min ago on a 60-min test → ~30 min left
		require.InDelta(t, 30*60, got.TimeRemainingSeconds, 5)
	})

	t.Run("S-07: auto-expires and returns status=expired when time is up", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), studentID, testID, 2*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 1)), nil)

		got, err := svc.Get(ctx, sess.ID, studentID)

		require.NoError(t, err)
		require.Equal(t, domain.SessionStatusExpired, got.Status)
		require.Equal(t, 0, got.TimeRemainingSeconds)
		require.Len(t, sessRepo.statuses, 1)
		require.Equal(t, domain.SessionStatusExpired, sessRepo.statuses[0].status)
		require.Equal(t, sess.ID, sessRepo.statuses[0].id)
	})

	t.Run("S-08: returns ErrForbidden when caller does not own the session", func(t *testing.T) {
		testID := uuid.New()
		ownerID, otherID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), ownerID, testID, 5*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 60)), nil)

		_, err := svc.Get(ctx, sess.ID, otherID)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("S-09: returns ErrNotFound when session does not exist", func(t *testing.T) {
		svc := NewService(newFakeSessionRepo(), &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.Get(ctx, uuid.New(), uuid.New())

		require.True(t, errors.Is(err, apierr.ErrNotFound))
	})
}

// ─── TestSaveAnswer ───────────────────────────────────────────────────────────

func TestSaveAnswer(t *testing.T) {
	ctx := context.Background()


	t.Run("S-10: upserts answer successfully for an active, in-time session", func(t *testing.T) {
		// SaveAnswer always calls questionInTestDB + questionType which need
		// a real pgxpool.Pool — not fakeable in a unit test.
		t.Skip("requires integration test: SaveAnswer uses pool for question validation")
	})

	t.Run("S-11: returns ErrConflict when session is already submitted", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := &domain.TestSession{
			ID: uuid.New(), StudentID: studentID, TestID: testID,
			StartedAt: time.Now().Add(-10 * time.Minute),
			Status:    domain.SessionStatusSubmitted,
		}
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 60)), nil)

		_, err := svc.SaveAnswer(ctx, sess.ID, studentID, AnswerInput{QuestionID: uuid.New()})

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("S-12: returns ErrConflict when session time has expired", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		// Session started 2 min ago on a 1-min test → overdue
		sess := makeInProgressSession(uuid.New(), studentID, testID, 2*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 1)), nil)

		_, err := svc.SaveAnswer(ctx, sess.ID, studentID, AnswerInput{QuestionID: uuid.New()})

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("S-13: returns ErrForbidden when caller does not own the session", func(t *testing.T) {
		ownerID, otherID, testID := uuid.New(), uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), ownerID, testID, 5*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(makePublishedTest(testID, 60)), nil)

		_, err := svc.SaveAnswer(ctx, sess.ID, otherID, AnswerInput{QuestionID: uuid.New()})

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	t.Run("S-14: returns ErrValidation when question is not in the test", func(t *testing.T) {
		// questionInTestDB hits *pgxpool.Pool directly — not fakeable in a unit test.
		t.Skip("requires integration test: questionInTestDB uses pgxpool.Pool")
	})

	t.Run("S-15: accepts nil selectedOptionID (blank answer)", func(t *testing.T) {
		// SaveAnswer always calls questionInTestDB + questionType which need
		// a real pgxpool.Pool — not fakeable in a unit test.
		t.Skip("requires integration test: SaveAnswer uses pool for question validation")
	})

	t.Run("S-16: upserts answer successfully and preserves selected option", func(t *testing.T) {
		// SaveAnswer always calls questionInTestDB + questionType which need
		// a real pgxpool.Pool — not fakeable in a unit test.
		t.Skip("requires integration test: SaveAnswer uses pool for question validation")
	})
}

// ─── TestToggleFlag ───────────────────────────────────────────────────────────

func TestToggleFlag(t *testing.T) {
	ctx := context.Background()

	t.Run("S-17: sets flag to true on an active session", func(t *testing.T) {
		studentID, testID, questionID := uuid.New(), uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), studentID, testID, 5*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.ToggleFlag(ctx, sess.ID, studentID, questionID, true)

		require.NoError(t, err)
		require.Len(t, sessRepo.flags, 1)
		require.Equal(t, questionID, sessRepo.flags[0].questionID)
		require.True(t, sessRepo.flags[0].flagged)
	})

	t.Run("S-18: clears flag (sets to false) on an active session", func(t *testing.T) {
		studentID, testID, questionID := uuid.New(), uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), studentID, testID, 5*time.Minute)
		sess.Answers = []domain.SessionAnswer{
			{ID: uuid.New(), SessionID: sess.ID, QuestionID: questionID, AnsweredAt: time.Now().UTC()},
		}
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.ToggleFlag(ctx, sess.ID, studentID, questionID, false)

		require.NoError(t, err)
		require.Len(t, sessRepo.flags, 1)
		require.False(t, sessRepo.flags[0].flagged)
	})

	t.Run("S-19: returns ErrConflict when session is not in_progress (submitted)", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := &domain.TestSession{
			ID: uuid.New(), StudentID: studentID, TestID: testID,
			StartedAt: time.Now().Add(-10 * time.Minute),
			Status:    domain.SessionStatusSubmitted,
		}
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.ToggleFlag(ctx, sess.ID, studentID, uuid.New(), true)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("S-20: returns ErrForbidden when caller does not own the session", func(t *testing.T) {
		ownerID, otherID, testID := uuid.New(), uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), ownerID, testID, 5*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.ToggleFlag(ctx, sess.ID, otherID, uuid.New(), true)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})
}

// ─── TestComputePercentageScore ──────────────────────────────────────────────

func TestComputePercentageScore(t *testing.T) {
	tests := []struct {
		name    string
		correct int
		total   int
		want    float64
	}{
		{"S-30: 3 correct out of 5 → 60.00", 3, 5, 60.00},
		{"S-31: 0 correct out of 5 → 0.00", 0, 5, 0.00},
		{"S-32: all correct → 100.00", 5, 5, 100.00},
		{"S-33: total=0 → 0.00 (no divide-by-zero)", 0, 0, 0.00},
		{"S-34: 2 correct out of 3 → 66.67", 2, 3, 66.67},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computePercentageScore(tc.correct, tc.total)
			require.InDelta(t, tc.want, got, 0.005)
		})
	}
}

// ─── TestComputeRemaining ─────────────────────────────────────────────────────

func TestComputeRemaining(t *testing.T) {
	t.Run("S-21: returns approximate seconds remaining for an in-time session", func(t *testing.T) {
		sess := &domain.TestSession{StartedAt: time.Now().Add(-30 * time.Minute)}

		got := computeRemaining(sess, 60)

		// 60-min test started 30 min ago → ~1800s left; allow 5s for test execution
		require.InDelta(t, 30*60, got, 5)
		require.Greater(t, got, 0)
	})

	t.Run("S-22: returns 0 (never negative) when session is past its deadline", func(t *testing.T) {
		sess := &domain.TestSession{StartedAt: time.Now().Add(-2 * time.Minute)}

		got := computeRemaining(sess, 1) // 1-min test started 2 min ago

		require.Equal(t, 0, got)
	})
}

// ─── TestSubmit ───────────────────────────────────────────────────────────────

func TestSubmit(t *testing.T) {
	ctx := context.Background()

	// S-25, S-26, S-27: error guards fire before the pool is touched — safe to unit test.

	t.Run("S-25: returns ErrConflict when session is already submitted", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := &domain.TestSession{
			ID:        uuid.New(),
			StudentID: studentID,
			TestID:    testID,
			Status:    domain.SessionStatusSubmitted,
		}
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.Submit(ctx, sess.ID, studentID)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("S-26: returns ErrConflict when session is expired", func(t *testing.T) {
		studentID, testID := uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := &domain.TestSession{
			ID:        uuid.New(),
			StudentID: studentID,
			TestID:    testID,
			Status:    domain.SessionStatusExpired,
		}
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.Submit(ctx, sess.ID, studentID)

		require.True(t, errors.Is(err, apierr.ErrConflict))
	})

	t.Run("S-27: returns ErrForbidden when caller does not own the session", func(t *testing.T) {
		ownerID, otherID, testID := uuid.New(), uuid.New(), uuid.New()
		sessRepo := newFakeSessionRepo()
		sess := makeInProgressSession(uuid.New(), ownerID, testID, 5*time.Minute)
		sessRepo.seed(sess)
		svc := NewService(sessRepo, &fakeResultRepo{}, newFakeTestRepo(), nil)

		_, err := svc.Submit(ctx, sess.ID, otherID)

		require.True(t, errors.Is(err, apierr.ErrForbidden))
	})

	// S-23, S-24, S-28, S-29 need a real Postgres connection.

	t.Run("S-23: scores correctly — correct/wrong/blank counts match formula (integration)", func(t *testing.T) {
		t.Skip("requires real Postgres: Submit calls pgstore.CorrectOptionMap and opens a pgx transaction")
	})

	t.Run("S-24: submit is atomic — session status and result row written together or not at all (integration)", func(t *testing.T) {
		t.Skip("requires real Postgres: atomicity cannot be verified without a real transaction")
	})

	t.Run("S-28: blank answer (no selection) scores as blank_points (integration)", func(t *testing.T) {
		t.Skip("requires real Postgres: Submit calls pgstore.CorrectOptionMap and opens a pgx transaction")
	})

	t.Run("S-29: negative wrong_points can reduce total_score below zero (integration)", func(t *testing.T) {
		t.Skip("requires real Postgres: Submit calls pgstore.CorrectOptionMap and opens a pgx transaction")
	})
}
