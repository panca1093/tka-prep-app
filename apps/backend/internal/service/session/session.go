package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
	pgstore "github.com/yourorg/tkaprep/apps/backend/internal/repository/postgres"
)

type Service struct {
	sessions repository.SessionRepository
	results  repository.ResultRepository
	tests    repository.TestRepository
	pool     *pgxpool.Pool
}

func NewService(
	sessions repository.SessionRepository,
	results repository.ResultRepository,
	tests repository.TestRepository,
	pool *pgxpool.Pool,
) *Service {
	return &Service{
		sessions: sessions,
		results:  results,
		tests:    tests,
		pool:     pool,
	}
}

// Start creates a new session for a student, or returns the existing in_progress one.
func (s *Service) Start(ctx context.Context, studentID, testID uuid.UUID) (*domain.TestSession, error) {
	t, err := s.tests.FindByID(ctx, testID)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TestStatusPublished {
		return nil, fmt.Errorf("%w: test is not published", apierr.ErrNotFound)
	}

	// Return existing active session rather than creating a duplicate.
	existing, err := s.sessions.FindActiveByStudentAndTest(ctx, studentID, testID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		remaining := computeRemaining(existing, t.DurationMinutes)
		if remaining <= 0 {
			now := time.Now().UTC()
			_ = s.sessions.UpdateStatus(ctx, existing.ID, domain.SessionStatusExpired, &now, 0)
			existing.Status = domain.SessionStatusExpired
		}
		if existing.Status == domain.SessionStatusInProgress {
			existing.TimeRemainingSeconds = remaining
			return existing, nil
		}
	}

	now := time.Now().UTC()
	sess := &domain.TestSession{
		ID:                   uuid.New(),
		StudentID:            studentID,
		TestID:               testID,
		StartedAt:            now,
		Status:               domain.SessionStatusInProgress,
		TimeRemainingSeconds: t.DurationMinutes * 60,
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	return sess, nil
}

// Get returns the session, refreshing time_remaining, auto-expiring if past deadline.
func (s *Service) Get(ctx context.Context, sessionID, callerID uuid.UUID) (*domain.TestSession, error) {
	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.StudentID != callerID {
		return nil, apierr.ErrForbidden
	}

	if sess.Status == domain.SessionStatusInProgress {
		t, err := s.tests.FindByID(ctx, sess.TestID)
		if err != nil {
			return nil, err
		}
		remaining := computeRemaining(sess, t.DurationMinutes)
		if remaining <= 0 {
			now := time.Now().UTC()
			_ = s.sessions.UpdateStatus(ctx, sess.ID, domain.SessionStatusExpired, &now, 0)
			sess.Status = domain.SessionStatusExpired
			sess.TimeRemainingSeconds = 0
		} else {
			sess.TimeRemainingSeconds = remaining
		}
	}
	return sess, nil
}

// SaveAnswer upserts an answer (blank if selectedOptionID is nil).
func (s *Service) SaveAnswer(ctx context.Context, sessionID, callerID, questionID uuid.UUID, selectedOptionID *uuid.UUID) (*domain.TestSession, error) {
	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.StudentID != callerID {
		return nil, apierr.ErrForbidden
	}
	if sess.Status != domain.SessionStatusInProgress {
		return nil, fmt.Errorf("%w: session is not in progress", apierr.ErrConflict)
	}

	t, err := s.tests.FindByID(ctx, sess.TestID)
	if err != nil {
		return nil, err
	}
	if computeRemaining(sess, t.DurationMinutes) <= 0 {
		now := time.Now().UTC()
		_ = s.sessions.UpdateStatus(ctx, sess.ID, domain.SessionStatusExpired, &now, 0)
		return nil, fmt.Errorf("%w: session has expired", apierr.ErrConflict)
	}

	// Verify question belongs to this test.
	if !questionInTest(sess, questionID) {
		if !questionInTestDB(ctx, s.pool, sess.TestID, questionID) {
			return nil, fmt.Errorf("%w: question not in this test", apierr.ErrValidation)
		}
	}

	a := &domain.SessionAnswer{
		ID:               uuid.New(),
		SessionID:        sessionID,
		QuestionID:       questionID,
		SelectedOptionID: selectedOptionID,
		AnsweredAt:       time.Now().UTC(),
	}
	// Preserve existing flag.
	for _, existing := range sess.Answers {
		if existing.QuestionID == questionID {
			a.IsFlagged = existing.IsFlagged
			break
		}
	}

	if err := s.sessions.UpsertAnswer(ctx, a); err != nil {
		return nil, err
	}

	// Reload to return fresh state.
	return s.sessions.FindByID(ctx, sessionID)
}

// ToggleFlag sets is_flagged for a question in the session.
func (s *Service) ToggleFlag(ctx context.Context, sessionID, callerID, questionID uuid.UUID, flagged bool) (*domain.TestSession, error) {
	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.StudentID != callerID {
		return nil, apierr.ErrForbidden
	}
	if sess.Status != domain.SessionStatusInProgress {
		return nil, fmt.Errorf("%w: session is not in progress", apierr.ErrConflict)
	}

	if err := s.sessions.UpsertFlag(ctx, sessionID, questionID, flagged); err != nil {
		return nil, err
	}
	return s.sessions.FindByID(ctx, sessionID)
}

// Submit finalizes the session and computes the result in one transaction.
func (s *Service) Submit(ctx context.Context, sessionID, callerID uuid.UUID) (*domain.TestResult, error) {
	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.StudentID != callerID {
		return nil, apierr.ErrForbidden
	}
	if sess.Status == domain.SessionStatusSubmitted {
		return nil, fmt.Errorf("%w: session already submitted", apierr.ErrConflict)
	}
	if sess.Status == domain.SessionStatusExpired {
		return nil, fmt.Errorf("%w: session has expired", apierr.ErrConflict)
	}

	t, err := s.tests.FindByID(ctx, sess.TestID)
	if err != nil {
		return nil, err
	}
	if t.ScoringConfig == nil {
		return nil, fmt.Errorf("test has no scoring config")
	}

	// Load correct-option map for all test questions.
	correctOpts, err := pgstore.CorrectOptionMap(ctx, s.pool, sess.TestID)
	if err != nil {
		return nil, err
	}

	// Build answer lookup.
	answered := make(map[uuid.UUID]*uuid.UUID)
	for _, a := range sess.Answers {
		optID := a.SelectedOptionID
		answered[a.QuestionID] = optID
	}

	// Count correct / wrong / blank across all questions in the test.
	correct, wrong, blank := 0, 0, 0
	for qID, correctOptID := range correctOpts {
		chosen, hasAnswer := answered[qID]
		if !hasAnswer || chosen == nil {
			blank++
		} else if *chosen == correctOptID {
			correct++
		} else {
			wrong++
		}
	}

	sc := t.ScoringConfig
	totalScore := (float64(correct) * sc.CorrectPoints) +
		(float64(wrong) * sc.WrongPoints) +
		(float64(blank) * sc.BlankPoints)

	now := time.Now().UTC()
	result := &domain.TestResult{
		ID:           uuid.New(),
		SessionID:    sessionID,
		StudentID:    callerID,
		TestID:       sess.TestID,
		TotalScore:   totalScore,
		CorrectCount: correct,
		WrongCount:   wrong,
		BlankCount:   blank,
		CompletedAt:  now,
	}

	// Transactional: update session status + insert result atomically.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx,
		`UPDATE test_sessions SET status='submitted', submitted_at=$1, time_remaining_seconds=0 WHERE id=$2`,
		now, sessionID,
	); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO test_results (id, session_id, student_id, test_id, total_score, correct_count, wrong_count, blank_count, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		result.ID, result.SessionID, result.StudentID, result.TestID,
		result.TotalScore, result.CorrectCount, result.WrongCount, result.BlankCount, result.CompletedAt,
	); err != nil {
		return nil, fmt.Errorf("insert result: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit submit: %w", err)
	}

	return result, nil
}

// computeRemaining calculates seconds remaining based on started_at vs now.
func computeRemaining(sess *domain.TestSession, durationMinutes int) int {
	elapsed := int(time.Since(sess.StartedAt).Seconds())
	total := durationMinutes * 60
	remaining := total - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func questionInTest(sess *domain.TestSession, questionID uuid.UUID) bool {
	for _, a := range sess.Answers {
		if a.QuestionID == questionID {
			return true
		}
	}
	return false
}

func questionInTestDB(ctx context.Context, pool *pgxpool.Pool, testID, questionID uuid.UUID) bool {
	var exists bool
	_ = pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM test_questions WHERE test_id=$1 AND question_id=$2)`,
		testID, questionID,
	).Scan(&exists)
	return exists
}
