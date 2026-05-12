package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
)

type ResultRepository struct {
	pool *pgxpool.Pool
}

func NewResultRepository(pool *pgxpool.Pool) *ResultRepository {
	return &ResultRepository{pool: pool}
}

func (r *ResultRepository) Create(ctx context.Context, res *domain.TestResult) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO test_results (id, session_id, student_id, test_id, total_score, correct_count, wrong_count, blank_count, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		res.ID, res.SessionID, res.StudentID, res.TestID,
		res.TotalScore, res.CorrectCount, res.WrongCount, res.BlankCount, res.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("create result: %w", err)
	}
	return nil
}

func (r *ResultRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) (*domain.TestResult, error) {
	res := &domain.TestResult{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, session_id, student_id, test_id, total_score, correct_count, wrong_count, blank_count, completed_at
		 FROM test_results WHERE session_id = $1`, sessionID,
	).Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID,
		&res.TotalScore, &res.CorrectCount, &res.WrongCount, &res.BlankCount, &res.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find result: %w", err)
	}
	return res, nil
}

func (r *ResultRepository) ListByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.TestResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, student_id, test_id, total_score, correct_count, wrong_count, blank_count, completed_at
		 FROM test_results WHERE student_id = $1 ORDER BY completed_at DESC`, studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}
	defer rows.Close()

	var results []*domain.TestResult
	for rows.Next() {
		res := &domain.TestResult{}
		if err := rows.Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID,
			&res.TotalScore, &res.CorrectCount, &res.WrongCount, &res.BlankCount, &res.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

func (r *ResultRepository) ListAll(ctx context.Context) ([]*domain.TestResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, student_id, test_id, total_score, correct_count, wrong_count, blank_count, completed_at
		 FROM test_results ORDER BY completed_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all results: %w", err)
	}
	defer rows.Close()

	var results []*domain.TestResult
	for rows.Next() {
		res := &domain.TestResult{}
		if err := rows.Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID,
			&res.TotalScore, &res.CorrectCount, &res.WrongCount, &res.BlankCount, &res.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

func (r *ResultRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.TestResult, error) {
	res := &domain.TestResult{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, session_id, student_id, test_id, total_score, correct_count, wrong_count, blank_count, completed_at
		 FROM test_results WHERE id = $1`, id,
	).Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID,
		&res.TotalScore, &res.CorrectCount, &res.WrongCount, &res.BlankCount, &res.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find result by id: %w", err)
	}
	return res, nil
}

func (r *ResultRepository) FindDetailByID(ctx context.Context, id uuid.UUID) (*domain.ResultDetail, error) {
	base, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Per-topic breakdown: for each topic referenced in the test, count correct/wrong/blank.
	rows, err := r.pool.Query(ctx,
		`SELECT
		     t.id, t.name,
		     COUNT(*)::int                                                        AS total,
		     SUM(CASE WHEN sa.selected_option_id IS NULL THEN 1 ELSE 0 END)::int AS blank_count,
		     SUM(CASE
		             WHEN sa.selected_option_id IS NOT NULL
		              AND sa.selected_option_id = co.id THEN 1
		             ELSE 0
		         END)::int AS correct_count,
		     SUM(CASE
		             WHEN sa.selected_option_id IS NOT NULL
		              AND sa.selected_option_id != co.id THEN 1
		             ELSE 0
		         END)::int AS wrong_count
		 FROM test_questions tq
		 JOIN questions q ON q.id = tq.question_id
		 JOIN topics t ON t.id = q.topic_id
		 JOIN question_options co ON co.question_id = q.id AND co.is_correct = true
		 LEFT JOIN session_answers sa ON sa.session_id = $1 AND sa.question_id = q.id
		 WHERE tq.test_id = $2
		 GROUP BY t.id, t.name
		 ORDER BY t.name`,
		base.SessionID, base.TestID,
	)
	if err != nil {
		return nil, fmt.Errorf("topic breakdown: %w", err)
	}
	defer rows.Close()

	detail := &domain.ResultDetail{TestResult: *base}
	total := base.CorrectCount + base.WrongCount + base.BlankCount
	if total > 0 {
		detail.Percentage = float64(base.CorrectCount) / float64(total) * 100
	}

	for rows.Next() {
		var tb domain.TopicBreakdown
		if err := rows.Scan(&tb.TopicID, &tb.TopicName, &tb.Total,
			&tb.BlankCount, &tb.CorrectCount, &tb.WrongCount); err != nil {
			return nil, fmt.Errorf("scan topic breakdown: %w", err)
		}
		detail.TopicBreakdown = append(detail.TopicBreakdown, tb)
	}
	return detail, rows.Err()
}

func (r *ResultRepository) GetReview(ctx context.Context, sessionID, testID uuid.UUID) ([]domain.ReviewItem, error) {
	// Load questions with their correct options and student answers.
	rows, err := r.pool.Query(ctx,
		`SELECT
		     q.id, tq.order_index, q.text, q.explanation, q.difficulty,
		     tp.id, tp.name,
		     sa.selected_option_id,
		     co.id AS correct_option_id
		 FROM test_questions tq
		 JOIN questions q ON q.id = tq.question_id
		 JOIN topics tp ON tp.id = q.topic_id
		 JOIN question_options co ON co.question_id = q.id AND co.is_correct = true
		 LEFT JOIN session_answers sa ON sa.session_id = $1 AND sa.question_id = q.id
		 WHERE tq.test_id = $2
		 ORDER BY tq.order_index`,
		sessionID, testID,
	)
	if err != nil {
		return nil, fmt.Errorf("review query: %w", err)
	}
	defer rows.Close()

	var items []domain.ReviewItem
	var questionIDs []uuid.UUID
	questionIndex := map[uuid.UUID]int{}

	for rows.Next() {
		var item domain.ReviewItem
		var diff string
		if err := rows.Scan(
			&item.QuestionID, &item.OrderIndex, &item.Text, &item.Explanation, &diff,
			&item.TopicID, &item.TopicName,
			&item.SelectedOptionID,
			&item.CorrectOptionID,
		); err != nil {
			return nil, fmt.Errorf("scan review item: %w", err)
		}
		item.Difficulty = domain.Difficulty(diff)
		if item.SelectedOptionID == nil {
			item.Status = domain.AnswerStatusBlank
		} else if *item.SelectedOptionID == item.CorrectOptionID {
			item.Status = domain.AnswerStatusCorrect
		} else {
			item.Status = domain.AnswerStatusWrong
		}
		questionIndex[item.QuestionID] = len(items)
		questionIDs = append(questionIDs, item.QuestionID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return items, nil
	}

	// Load all options for these questions in one query, then attach.
	optRows, err := r.pool.Query(ctx,
		`SELECT id, question_id, label, text, is_correct
		 FROM question_options
		 WHERE question_id = ANY($1)
		 ORDER BY label`, questionIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load review options: %w", err)
	}
	defer optRows.Close()

	for optRows.Next() {
		var opt domain.ReviewOption
		var qID uuid.UUID
		if err := optRows.Scan(&opt.ID, &qID, &opt.Label, &opt.Text, &opt.IsCorrect); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		if idx, ok := questionIndex[qID]; ok {
			items[idx].Options = append(items[idx].Options, opt)
		}
	}
	return items, optRows.Err()
}
