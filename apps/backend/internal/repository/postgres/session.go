package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.TestSession) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO test_sessions (id, student_id, test_id, started_at, status, time_remaining_seconds)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		s.ID, s.StudentID, s.TestID, s.StartedAt, string(s.Status), s.TimeRemainingSeconds,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.TestSession, error) {
	s := &domain.TestSession{}
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, test_id, started_at, submitted_at, status, time_remaining_seconds
		 FROM test_sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.StudentID, &s.TestID, &s.StartedAt, &s.SubmittedAt, &status, &s.TimeRemainingSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find session: %w", err)
	}
	s.Status = domain.SessionStatus(status)
	if err := r.loadAnswersAndFlags(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SessionRepository) FindActiveByStudentAndTest(ctx context.Context, studentID, testID uuid.UUID) (*domain.TestSession, error) {
	s := &domain.TestSession{}
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, test_id, started_at, submitted_at, status, time_remaining_seconds
		 FROM test_sessions WHERE student_id = $1 AND test_id = $2 AND status = 'in_progress'
		 ORDER BY started_at DESC LIMIT 1`, studentID, testID,
	).Scan(&s.ID, &s.StudentID, &s.TestID, &s.StartedAt, &s.SubmittedAt, &status, &s.TimeRemainingSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find active session: %w", err)
	}
	s.Status = domain.SessionStatus(status)
	return s, nil
}

func (r *SessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SessionStatus, submittedAt *time.Time, timeRemaining int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE test_sessions SET status=$1, submitted_at=$2, time_remaining_seconds=$3 WHERE id=$4`,
		string(status), submittedAt, timeRemaining, id,
	)
	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

// SaveMCQAnswer replaces the single MCQ answer for a question.
func (r *SessionRepository) SaveMCQAnswer(ctx context.Context, sessionID, questionID uuid.UUID, optionID *uuid.UUID, now time.Time) error {
	// Delete existing answer for this question then insert the new one.
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM session_answers WHERE session_id=$1 AND question_id=$2`, sessionID, questionID,
	); err != nil {
		return fmt.Errorf("delete old mcq answer: %w", err)
	}
	if optionID == nil {
		return nil // blank answer — just deleted
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO session_answers (id, session_id, question_id, selected_option_id, answered_at)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (session_id, question_id, selected_option_id) WHERE selected_option_id IS NOT NULL
		 DO UPDATE SET answered_at = EXCLUDED.answered_at`,
		uuid.New(), sessionID, questionID, optionID, now,
	)
	if err != nil {
		return fmt.Errorf("save mcq answer: %w", err)
	}
	return nil
}

// SavePGKAnswers replaces the selected-option set for a multi_correct question.
func (r *SessionRepository) SavePGKAnswers(ctx context.Context, sessionID, questionID uuid.UUID, optionIDs []uuid.UUID, now time.Time) error {
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM session_answers WHERE session_id=$1 AND question_id=$2`, sessionID, questionID,
	); err != nil {
		return fmt.Errorf("delete old pgk answers: %w", err)
	}
	for _, oid := range optionIDs {
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO session_answers (id, session_id, question_id, selected_option_id, answered_at)
			 VALUES ($1,$2,$3,$4,$5)`,
			uuid.New(), sessionID, questionID, oid, now,
		); err != nil {
			return fmt.Errorf("insert pgk answer: %w", err)
		}
	}
	return nil
}

// SaveBSAnswers replaces the statement answers for a true_false question.
func (r *SessionRepository) SaveBSAnswers(ctx context.Context, sessionID, questionID uuid.UUID, stmtAnswers []domain.StatementAnswerInput, now time.Time) error {
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM session_answers WHERE session_id=$1 AND question_id=$2`, sessionID, questionID,
	); err != nil {
		return fmt.Errorf("delete old bs answers: %w", err)
	}
	for _, sa := range stmtAnswers {
		b := sa.IsCorrect
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO session_answers (id, session_id, question_id, statement_id, boolean_answer, answered_at)
			 VALUES ($1,$2,$3,$4,$5,$6)`,
			uuid.New(), sessionID, questionID, sa.StatementID, b, now,
		); err != nil {
			return fmt.Errorf("insert bs answer: %w", err)
		}
	}
	return nil
}

// UpsertFlag sets the flag on a question in this session.
func (r *SessionRepository) UpsertFlag(ctx context.Context, sessionID, questionID uuid.UUID, flagged bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO session_question_flags (session_id, question_id, is_flagged)
		 VALUES ($1,$2,$3)
		 ON CONFLICT (session_id, question_id)
		 DO UPDATE SET is_flagged = EXCLUDED.is_flagged`,
		sessionID, questionID, flagged,
	)
	if err != nil {
		return fmt.Errorf("upsert flag: %w", err)
	}
	return nil
}

func (r *SessionRepository) loadAnswersAndFlags(ctx context.Context, s *domain.TestSession) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, question_id, selected_option_id, statement_id, boolean_answer, answered_at
		 FROM session_answers WHERE session_id = $1`, s.ID,
	)
	if err != nil {
		return fmt.Errorf("load answers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a domain.SessionAnswer
		if err := rows.Scan(&a.ID, &a.SessionID, &a.QuestionID,
			&a.SelectedOptionID, &a.StatementID, &a.BooleanAnswer, &a.AnsweredAt); err != nil {
			return fmt.Errorf("scan answer: %w", err)
		}
		s.Answers = append(s.Answers, a)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	flagRows, err := r.pool.Query(ctx,
		`SELECT question_id, is_flagged FROM session_question_flags WHERE session_id = $1`, s.ID,
	)
	if err != nil {
		return fmt.Errorf("load flags: %w", err)
	}
	defer flagRows.Close()

	for flagRows.Next() {
		var f domain.SessionQuestionFlag
		if err := flagRows.Scan(&f.QuestionID, &f.IsFlagged); err != nil {
			return fmt.Errorf("scan flag: %w", err)
		}
		s.Flags = append(s.Flags, f)
	}
	return flagRows.Err()
}

// TestQuestionInfo holds the type of a question inside a test.
type TestQuestionInfo struct {
	QuestionID   uuid.UUID
	QuestionType domain.QuestionType
}

// TestQuestionsInfo loads question IDs + types for all questions in a test.
func TestQuestionsInfo(ctx context.Context, pool *pgxpool.Pool, testID uuid.UUID) ([]TestQuestionInfo, error) {
	rows, err := pool.Query(ctx,
		`SELECT tq.question_id, q.question_type
		 FROM test_questions tq
		 JOIN questions q ON q.id = tq.question_id
		 WHERE tq.test_id = $1
		 ORDER BY tq.order_index`, testID,
	)
	if err != nil {
		return nil, fmt.Errorf("load test question types: %w", err)
	}
	defer rows.Close()

	var out []TestQuestionInfo
	for rows.Next() {
		var info TestQuestionInfo
		var qt string
		if err := rows.Scan(&info.QuestionID, &qt); err != nil {
			return nil, fmt.Errorf("scan question info: %w", err)
		}
		info.QuestionType = domain.QuestionType(qt)
		out = append(out, info)
	}
	return out, rows.Err()
}

// CorrectAnswerMap loads scoring data for all questions in a test.
// Returns: map[questionID] → ScoringData.
type ScoringData struct {
	Type             domain.QuestionType
	CorrectOptionIDs []uuid.UUID            // MCQ: 1 entry; PGK: N entries
	Statements       []domain.QuestionStatement // true_false
}

func LoadScoringData(ctx context.Context, pool *pgxpool.Pool, testID uuid.UUID) (map[uuid.UUID]*ScoringData, error) {
	// Load question types.
	infos, err := TestQuestionsInfo(ctx, pool, testID)
	if err != nil {
		return nil, err
	}

	m := make(map[uuid.UUID]*ScoringData, len(infos))
	for _, info := range infos {
		m[info.QuestionID] = &ScoringData{Type: info.QuestionType}
	}

	// Load correct options for MCQ / multi_correct.
	optRows, err := pool.Query(ctx,
		`SELECT tq.question_id, qo.id
		 FROM test_questions tq
		 JOIN question_options qo ON qo.question_id = tq.question_id AND qo.is_correct = true
		 WHERE tq.test_id = $1`, testID,
	)
	if err != nil {
		return nil, fmt.Errorf("load correct options: %w", err)
	}
	defer optRows.Close()

	for optRows.Next() {
		var qID, optID uuid.UUID
		if err := optRows.Scan(&qID, &optID); err != nil {
			return nil, fmt.Errorf("scan correct option: %w", err)
		}
		if sd, ok := m[qID]; ok {
			sd.CorrectOptionIDs = append(sd.CorrectOptionIDs, optID)
		}
	}
	if err := optRows.Err(); err != nil {
		return nil, err
	}

	// Load statements for true_false questions.
	stmtRows, err := pool.Query(ctx,
		`SELECT tq.question_id, qs.id, qs.text, qs.is_correct, qs.position
		 FROM test_questions tq
		 JOIN question_statements qs ON qs.question_id = tq.question_id
		 WHERE tq.test_id = $1
		 ORDER BY qs.position`, testID,
	)
	if err != nil {
		return nil, fmt.Errorf("load statements: %w", err)
	}
	defer stmtRows.Close()

	for stmtRows.Next() {
		var qID uuid.UUID
		var s domain.QuestionStatement
		if err := stmtRows.Scan(&qID, &s.ID, &s.Text, &s.IsCorrect, &s.Position); err != nil {
			return nil, fmt.Errorf("scan statement: %w", err)
		}
		if sd, ok := m[qID]; ok {
			sd.Statements = append(sd.Statements, s)
		}
	}
	return m, stmtRows.Err()
}
