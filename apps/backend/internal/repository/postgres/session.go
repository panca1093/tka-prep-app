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

	if err := r.LoadAnswers(ctx, s); err != nil {
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

func (r *SessionRepository) UpsertAnswer(ctx context.Context, a *domain.SessionAnswer) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO session_answers (id, session_id, question_id, selected_option_id, is_flagged, answered_at)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (session_id, question_id)
		 DO UPDATE SET selected_option_id = EXCLUDED.selected_option_id,
		               answered_at = EXCLUDED.answered_at`,
		a.ID, a.SessionID, a.QuestionID, a.SelectedOptionID, a.IsFlagged, a.AnsweredAt,
	)
	if err != nil {
		return fmt.Errorf("upsert answer: %w", err)
	}
	return nil
}

func (r *SessionRepository) UpsertFlag(ctx context.Context, sessionID, questionID uuid.UUID, flagged bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO session_answers (id, session_id, question_id, is_flagged, answered_at)
		 VALUES ($1,$2,$3,$4,now())
		 ON CONFLICT (session_id, question_id)
		 DO UPDATE SET is_flagged = EXCLUDED.is_flagged`,
		uuid.New(), sessionID, questionID, flagged,
	)
	if err != nil {
		return fmt.Errorf("upsert flag: %w", err)
	}
	return nil
}

func (r *SessionRepository) LoadAnswers(ctx context.Context, s *domain.TestSession) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, question_id, selected_option_id, is_flagged, answered_at
		 FROM session_answers WHERE session_id = $1`, s.ID,
	)
	if err != nil {
		return fmt.Errorf("load answers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a domain.SessionAnswer
		if err := rows.Scan(&a.ID, &a.SessionID, &a.QuestionID, &a.SelectedOptionID, &a.IsFlagged, &a.AnsweredAt); err != nil {
			return fmt.Errorf("scan answer: %w", err)
		}
		s.Answers = append(s.Answers, a)
	}
	return rows.Err()
}

// CorrectOptionMap loads a map of question_id → correct option_id for all questions in a session's test.
func CorrectOptionMap(ctx context.Context, pool *pgxpool.Pool, testID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	rows, err := pool.Query(ctx,
		`SELECT tq.question_id, qo.id
		 FROM test_questions tq
		 JOIN question_options qo ON qo.question_id = tq.question_id AND qo.is_correct = true
		 WHERE tq.test_id = $1`, testID,
	)
	if err != nil {
		return nil, fmt.Errorf("load correct options: %w", err)
	}
	defer rows.Close()

	m := make(map[uuid.UUID]uuid.UUID)
	for rows.Next() {
		var qID, optID uuid.UUID
		if err := rows.Scan(&qID, &optID); err != nil {
			return nil, fmt.Errorf("scan correct option: %w", err)
		}
		m[qID] = optID
	}
	return m, rows.Err()
}
