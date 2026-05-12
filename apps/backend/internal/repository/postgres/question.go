package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type QuestionRepository struct {
	pool *pgxpool.Pool
}

func NewQuestionRepository(pool *pgxpool.Pool) *QuestionRepository {
	return &QuestionRepository{pool: pool}
}

func (r *QuestionRepository) List(ctx context.Context, f repository.QuestionFilter) ([]*domain.Question, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	args := []any{}
	conds := []string{}
	i := 1

	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("q.text ILIKE $%d", i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	if f.TopicID != nil {
		conds = append(conds, fmt.Sprintf("q.topic_id = $%d", i))
		args = append(args, *f.TopicID)
		i++
	}
	if f.Difficulty != nil {
		conds = append(conds, fmt.Sprintf("q.difficulty = $%d", i))
		args = append(args, string(*f.Difficulty))
		i++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM questions q %s`, where),
		countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count questions: %w", err)
	}

	args = append(args, f.Limit, offset)
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT q.id, q.contributor_id, q.topic_id, q.text, q.explanation,
		       q.difficulty, q.created_at, q.updated_at
		  FROM questions q %s
		  ORDER BY q.created_at DESC
		  LIMIT $%d OFFSET $%d`, where, i, i+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list questions: %w", err)
	}
	defer rows.Close()

	var qs []*domain.Question
	for rows.Next() {
		q := &domain.Question{}
		var diff string
		if err := rows.Scan(&q.ID, &q.ContributorID, &q.TopicID, &q.Text, &q.Explanation,
			&diff, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan question: %w", err)
		}
		q.Difficulty = domain.Difficulty(diff)
		qs = append(qs, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load options for each question.
	for _, q := range qs {
		if err := r.loadOptions(ctx, q); err != nil {
			return nil, 0, err
		}
	}
	return qs, total, nil
}

func (r *QuestionRepository) Create(ctx context.Context, q *domain.Question) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO questions (id, contributor_id, topic_id, text, explanation, difficulty, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		q.ID, q.ContributorID, q.TopicID, q.Text, q.Explanation,
		string(q.Difficulty), q.CreatedAt, q.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert question: %w", err)
	}

	if err := insertOptions(ctx, tx, q.ID, q.Options); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *QuestionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	q := &domain.Question{}
	var diff string
	err := r.pool.QueryRow(ctx,
		`SELECT id, contributor_id, topic_id, text, explanation, difficulty, created_at, updated_at
		 FROM questions WHERE id = $1`, id,
	).Scan(&q.ID, &q.ContributorID, &q.TopicID, &q.Text, &q.Explanation,
		&diff, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find question by id: %w", err)
	}
	q.Difficulty = domain.Difficulty(diff)
	if err := r.loadOptions(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (r *QuestionRepository) Update(ctx context.Context, q *domain.Question) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	tag, err := tx.Exec(ctx,
		`UPDATE questions SET topic_id=$1, text=$2, explanation=$3, difficulty=$4, updated_at=$5
		 WHERE id=$6`,
		q.TopicID, q.Text, q.Explanation, string(q.Difficulty), q.UpdatedAt, q.ID,
	)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}

	// Replace options if provided.
	if len(q.Options) > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM question_options WHERE question_id = $1`, q.ID); err != nil {
			return fmt.Errorf("delete old options: %w", err)
		}
		if err := insertOptions(ctx, tx, q.ID, q.Options); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *QuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM questions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *QuestionRepository) IsUsedInPublishedTest(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM test_questions tq
		 JOIN tests t ON t.id = tq.test_id
		 WHERE tq.question_id = $1 AND t.status = 'published'
		 LIMIT 1`, id,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check question usage: %w", err)
	}
	return count > 0, nil
}

// loadOptions fetches options for a question and attaches them.
func (r *QuestionRepository) loadOptions(ctx context.Context, q *domain.Question) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, question_id, label, text, is_correct
		 FROM question_options WHERE question_id = $1 ORDER BY label`, q.ID,
	)
	if err != nil {
		return fmt.Errorf("load options: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var opt domain.QuestionOption
		if err := rows.Scan(&opt.ID, &opt.QuestionID, &opt.Label, &opt.Text, &opt.IsCorrect); err != nil {
			return fmt.Errorf("scan option: %w", err)
		}
		q.Options = append(q.Options, opt)
	}
	return rows.Err()
}

// insertOptions bulk-inserts question options inside the given transaction.
func insertOptions(ctx context.Context, tx pgx.Tx, questionID uuid.UUID, opts []domain.QuestionOption) error {
	for _, opt := range opts {
		_, err := tx.Exec(ctx,
			`INSERT INTO question_options (id, question_id, label, text, is_correct)
			 VALUES ($1, $2, $3, $4, $5)`,
			opt.ID, questionID, opt.Label, opt.Text, opt.IsCorrect,
		)
		if err != nil {
			return fmt.Errorf("insert option %s: %w", opt.Label, err)
		}
	}
	return nil
}
