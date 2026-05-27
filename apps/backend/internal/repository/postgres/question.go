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
	if f.QuestionType != nil {
		conds = append(conds, fmt.Sprintf("q.question_type = $%d", i))
		args = append(args, string(*f.QuestionType))
		i++
	}
	if f.EducationLevel != nil {
		conds = append(conds, fmt.Sprintf("q.education_level = $%d", i))
		args = append(args, *f.EducationLevel)
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
		fmt.Sprintf(`SELECT q.id, q.contributor_id, u_owner.name, q.topic_id, q.question_type, q.education_level, q.text, q.explanation,
		       q.image_url, q.difficulty, q.created_at, q.updated_at
		  FROM questions q
		  JOIN users u_owner ON u_owner.id = q.contributor_id
		  %s
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
		var diff, qtype string
		if err := rows.Scan(&q.ID, &q.ContributorID, &q.ContributorName, &q.TopicID, &qtype, &q.EducationLevel, &q.Text, &q.Explanation,
			&q.ImageURL, &diff, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan question: %w", err)
		}
		q.Difficulty = domain.Difficulty(diff)
		q.Type = domain.QuestionType(qtype)

		// Populate usage stats only for the question owner.
		if f.CallerID != uuid.Nil && q.ContributorID == f.CallerID {
			if err := r.loadUsageStats(ctx, q); err != nil {
				return nil, 0, err
			}
		}
		qs = append(qs, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for _, q := range qs {
		if err := r.loadOptionsAndStatements(ctx, q); err != nil {
			return nil, 0, err
		}
	}
	return qs, total, nil
}

func (r *QuestionRepository) loadUsageStats(ctx context.Context, q *domain.Question) error {
	stats := &domain.QuestionUsageStats{}
	err := r.pool.QueryRow(ctx,
		`SELECT
			COUNT(DISTINCT CASE WHEN t.contributor_id = $2 THEN tq.test_id END)::int,
			COUNT(DISTINCT CASE WHEN t.contributor_id != $2 THEN tq.test_id END)::int
		 FROM test_questions tq
		 JOIN tests t ON t.id = tq.test_id AND t.status = 'published'
		 WHERE tq.question_id = $1`, q.ID, q.ContributorID,
	).Scan(&stats.OwnTestCount, &stats.OtherTestCount)
	if err != nil {
		return fmt.Errorf("load usage stats: %w", err)
	}
	q.UsageStats = stats
	return nil
}

func (r *QuestionRepository) Create(ctx context.Context, q *domain.Question) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO questions (id, contributor_id, topic_id, question_type, education_level, text, explanation, image_url, difficulty, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		q.ID, q.ContributorID, q.TopicID, string(q.Type), q.EducationLevel, q.Text, q.Explanation, q.ImageURL,
		string(q.Difficulty), q.CreatedAt, q.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert question: %w", err)
	}

	if len(q.Options) > 0 {
		if err := insertOptions(ctx, tx, q.ID, q.Options); err != nil {
			return err
		}
	}
	if len(q.Statements) > 0 {
		if err := insertStatements(ctx, tx, q.ID, q.Statements); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *QuestionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	q := &domain.Question{}
	var diff, qtype string
	err := r.pool.QueryRow(ctx,
		`SELECT id, contributor_id, topic_id, question_type, education_level, text, explanation, image_url, difficulty, created_at, updated_at
		 FROM questions WHERE id = $1`, id,
	).Scan(&q.ID, &q.ContributorID, &q.TopicID, &qtype, &q.EducationLevel, &q.Text, &q.Explanation,
		&q.ImageURL, &diff, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find question by id: %w", err)
	}
	q.Difficulty = domain.Difficulty(diff)
	q.Type = domain.QuestionType(qtype)
	if err := r.loadOptionsAndStatements(ctx, q); err != nil {
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
		`UPDATE questions SET topic_id=$1, education_level=$2, text=$3, explanation=$4, image_url=$5, difficulty=$6, updated_at=$7
		 WHERE id=$8`,
		q.TopicID, q.EducationLevel, q.Text, q.Explanation, q.ImageURL, string(q.Difficulty), q.UpdatedAt, q.ID,
	)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}

	if len(q.Options) > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM question_options WHERE question_id = $1`, q.ID); err != nil {
			return fmt.Errorf("delete old options: %w", err)
		}
		if err := insertOptions(ctx, tx, q.ID, q.Options); err != nil {
			return err
		}
	}
	if len(q.Statements) > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM question_statements WHERE question_id = $1`, q.ID); err != nil {
			return fmt.Errorf("delete old statements: %w", err)
		}
		if err := insertStatements(ctx, tx, q.ID, q.Statements); err != nil {
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

func (r *QuestionRepository) loadOptionsAndStatements(ctx context.Context, q *domain.Question) error {
	if q.Type != domain.QuestionTypeTrueFalse {
		return r.loadOptions(ctx, q)
	}
	return r.loadStatements(ctx, q)
}

func (r *QuestionRepository) loadOptions(ctx context.Context, q *domain.Question) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, question_id, label, text, is_correct, image_url
		 FROM question_options WHERE question_id = $1 ORDER BY label`, q.ID,
	)
	if err != nil {
		return fmt.Errorf("load options: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var opt domain.QuestionOption
		if err := rows.Scan(&opt.ID, &opt.QuestionID, &opt.Label, &opt.Text, &opt.IsCorrect, &opt.ImageURL); err != nil {
			return fmt.Errorf("scan option: %w", err)
		}
		q.Options = append(q.Options, opt)
	}
	return rows.Err()
}

func (r *QuestionRepository) loadStatements(ctx context.Context, q *domain.Question) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, question_id, text, is_correct, position, image_url
		 FROM question_statements WHERE question_id = $1 ORDER BY position`, q.ID,
	)
	if err != nil {
		return fmt.Errorf("load statements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s domain.QuestionStatement
		if err := rows.Scan(&s.ID, &s.QuestionID, &s.Text, &s.IsCorrect, &s.Position, &s.ImageURL); err != nil {
			return fmt.Errorf("scan statement: %w", err)
		}
		q.Statements = append(q.Statements, s)
	}
	return rows.Err()
}

func insertOptions(ctx context.Context, tx pgx.Tx, questionID uuid.UUID, opts []domain.QuestionOption) error {
	for _, opt := range opts {
		_, err := tx.Exec(ctx,
			`INSERT INTO question_options (id, question_id, label, text, is_correct, image_url)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			opt.ID, questionID, opt.Label, opt.Text, opt.IsCorrect, opt.ImageURL,
		)
		if err != nil {
			return fmt.Errorf("insert option %s: %w", opt.Label, err)
		}
	}
	return nil
}

func insertStatements(ctx context.Context, tx pgx.Tx, questionID uuid.UUID, stmts []domain.QuestionStatement) error {
	for _, s := range stmts {
		_, err := tx.Exec(ctx,
			`INSERT INTO question_statements (id, question_id, text, is_correct, position, image_url)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			s.ID, questionID, s.Text, s.IsCorrect, s.Position, s.ImageURL,
		)
		if err != nil {
			return fmt.Errorf("insert statement: %w", err)
		}
	}
	return nil
}

func (r *QuestionRepository) ListUsageStats(ctx context.Context, contributorID uuid.UUID, limit int) ([]domain.QuestionUsageEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx,
		`SELECT q.id, LEFT(q.text, 100), tp.name,
		        COUNT(DISTINCT CASE WHEN t.contributor_id = q.contributor_id THEN tq.test_id END)::int,
		        COUNT(DISTINCT CASE WHEN t.contributor_id != q.contributor_id THEN tq.test_id END)::int
		 FROM questions q
		 JOIN topics tp ON tp.id = q.topic_id
		 JOIN test_questions tq ON tq.question_id = q.id
		 JOIN tests t ON t.id = tq.test_id AND t.status = 'published'
		 WHERE q.contributor_id = $1
		 GROUP BY q.id, tp.name
		 ORDER BY (COUNT(DISTINCT CASE WHEN t.contributor_id = q.contributor_id THEN tq.test_id END) +
		           COUNT(DISTINCT CASE WHEN t.contributor_id != q.contributor_id THEN tq.test_id END)) DESC
		 LIMIT $2`,
		contributorID, limit)
	if err != nil {
		return nil, fmt.Errorf("list usage stats: %w", err)
	}
	defer rows.Close()

	var out []domain.QuestionUsageEntry
	for rows.Next() {
		var e domain.QuestionUsageEntry
		if err := rows.Scan(&e.QuestionID, &e.QuestionText, &e.TopicName, &e.OwnTestCount, &e.OtherTestCount); err != nil {
			return nil, fmt.Errorf("scan usage entry: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
