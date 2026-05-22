package session

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
	pgstore "github.com/yourorg/tkaprep/apps/backend/internal/repository/postgres"
)

var _ = pgstore.LoadScoringData // ensure pgstore import is used

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
	return &Service{sessions: sessions, results: results, tests: tests, pool: pool}
}

// AnswerInput is the parsed answer payload from the handler.
type AnswerInput struct {
	QuestionID      uuid.UUID
	// MCQ
	SelectedOptionID *uuid.UUID
	// PGK
	SelectedOptionIDs []uuid.UUID
	// B/S
	StatementAnswers []domain.StatementAnswerInput
}

func (s *Service) Start(ctx context.Context, studentID, testID uuid.UUID) (*domain.TestSession, error) {
	t, err := s.tests.FindByID(ctx, testID)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TestStatusPublished {
		return nil, fmt.Errorf("%w: test is not published", apierr.ErrNotFound)
	}

	// Guard against empty tests — should not happen with publish validation
	// but protects against stale data or direct DB manipulation.
	if len(t.Questions) == 0 {
		return nil, fmt.Errorf("%w: test has no questions", apierr.ErrValidation)
	}

	existing, err := s.sessions.FindActiveByStudentAndTest(ctx, studentID, testID)
	if err != nil {
		return nil, err
	}
	// Prevent re-taking: check if student already has a completed result.
	var alreadyCompleted bool
	_ = s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM test_results WHERE student_id = $1 AND test_id = $2)`,
		studentID, testID,
	).Scan(&alreadyCompleted)
	if alreadyCompleted {
		return nil, fmt.Errorf("%w: test already completed", apierr.ErrConflict)
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

func (s *Service) SaveAnswer(ctx context.Context, sessionID, callerID uuid.UUID, in AnswerInput) (*domain.TestSession, error) {
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

	if !questionInTestDB(ctx, s.pool, sess.TestID, in.QuestionID) {
		return nil, fmt.Errorf("%w: question not in this test", apierr.ErrValidation)
	}

	// Load question type to route to correct save method.
	qtype, err := questionType(ctx, s.pool, in.QuestionID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	switch qtype {
	case domain.QuestionTypeMCQ:
		if err := s.sessions.SaveMCQAnswer(ctx, sessionID, in.QuestionID, in.SelectedOptionID, now); err != nil {
			return nil, err
		}
	case domain.QuestionTypeMultiCorrect:
		if err := s.sessions.SavePGKAnswers(ctx, sessionID, in.QuestionID, in.SelectedOptionIDs, now); err != nil {
			return nil, err
		}
	case domain.QuestionTypeTrueFalse:
		if err := s.sessions.SaveBSAnswers(ctx, sessionID, in.QuestionID, in.StatementAnswers, now); err != nil {
			return nil, err
		}
	}

	return s.sessions.FindByID(ctx, sessionID)
}

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

	scoringData, err := pgstore.LoadScoringData(ctx, s.pool, sess.TestID)
	if err != nil {
		return nil, err
	}

	// Group student answers by question.
	type answerGroup struct {
		optionIDs   []uuid.UUID
		stmtAnswers map[uuid.UUID]bool // statementID → student's boolean_answer
	}
	groups := make(map[uuid.UUID]*answerGroup)
	for _, a := range sess.Answers {
		g, ok := groups[a.QuestionID]
		if !ok {
			g = &answerGroup{stmtAnswers: make(map[uuid.UUID]bool)}
			groups[a.QuestionID] = g
		}
		if a.SelectedOptionID != nil {
			g.optionIDs = append(g.optionIDs, *a.SelectedOptionID)
		}
		if a.StatementID != nil && a.BooleanAnswer != nil {
			g.stmtAnswers[*a.StatementID] = *a.BooleanAnswer
		}
	}

	sc := t.ScoringConfig
	totalScore := 0.0
	correct, wrong, blank := 0, 0, 0

	for qID, sd := range scoringData {
		g := groups[qID]
		switch sd.Type {
		case domain.QuestionTypeMCQ:
			if g == nil || len(g.optionIDs) == 0 {
				blank++
				totalScore += sc.BlankPoints
			} else if len(sd.CorrectOptionIDs) > 0 && g.optionIDs[0] == sd.CorrectOptionIDs[0] {
				correct++
				totalScore += sc.CorrectPoints
			} else {
				wrong++
				totalScore += sc.WrongPoints
			}

		case domain.QuestionTypeMultiCorrect:
			if g == nil || len(g.optionIDs) == 0 {
				blank++
				totalScore += sc.BlankPoints
			} else if setsEqual(g.optionIDs, sd.CorrectOptionIDs) {
				correct++
				totalScore += sc.CorrectPoints
			} else {
				wrong++
				totalScore += sc.WrongPoints
			}

		case domain.QuestionTypeTrueFalse:
			if g == nil || len(g.stmtAnswers) == 0 {
				blank++
				totalScore += sc.BlankPoints
			} else {
				total := len(sd.Statements)
				correctStmts := 0
				for _, stmt := range sd.Statements {
					if ans, answered := g.stmtAnswers[stmt.ID]; answered && ans == stmt.IsCorrect {
						correctStmts++
					}
				}
				proportion := float64(correctStmts) / float64(total)
				qScore := proportion * sc.CorrectPoints
				totalScore += qScore
				if correctStmts == total {
					correct++
				} else {
					wrong++
				}
			}
		}
	}

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

	return result, tx.Commit(ctx)
}

func computeRemaining(sess *domain.TestSession, durationMinutes int) int {
	elapsed := int(time.Since(sess.StartedAt).Seconds())
	total := durationMinutes * 60
	if r := total - elapsed; r > 0 {
		return r
	}
	return 0
}

func questionInTestDB(ctx context.Context, pool *pgxpool.Pool, testID, questionID uuid.UUID) bool {
	var exists bool
	_ = pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM test_questions WHERE test_id=$1 AND question_id=$2)`,
		testID, questionID,
	).Scan(&exists)
	return exists
}

func questionType(ctx context.Context, pool *pgxpool.Pool, questionID uuid.UUID) (domain.QuestionType, error) {
	var qt string
	err := pool.QueryRow(ctx, `SELECT question_type FROM questions WHERE id=$1`, questionID).Scan(&qt)
	if err != nil {
		return "", fmt.Errorf("get question type: %w", err)
	}
	return domain.QuestionType(qt), nil
}

func setsEqual(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	toKey := func(ids []uuid.UUID) string {
		strs := make([]string, len(ids))
		for i, id := range ids {
			strs[i] = id.String()
		}
		sort.Strings(strs)
		out := ""
		for _, s := range strs {
			out += s
		}
		return out
	}
	return toKey(a) == toKey(b)
}
