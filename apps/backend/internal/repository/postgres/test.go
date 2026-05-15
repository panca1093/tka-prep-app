package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type TestRepository struct {
	pool *pgxpool.Pool
}

func NewTestRepository(pool *pgxpool.Pool) *TestRepository {
	return &TestRepository{pool: pool}
}

// Create inserts the test and its scoring config in one transaction.
func (r *TestRepository) Create(ctx context.Context, t *domain.Test) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO tests (id, contributor_id, title, description, category, duration_minutes, difficulty, status, education_level, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.ContributorID, t.Title, t.Description,
		string(t.Category), t.DurationMinutes, string(t.Difficulty),
		string(t.Status), t.EducationLevel, t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert test: %w", err)
	}

	if t.ScoringConfig != nil {
		_, err = tx.Exec(ctx,
			`INSERT INTO scoring_configs (id, test_id, correct_points, wrong_points, blank_points)
			 VALUES ($1,$2,$3,$4,$5)`,
			t.ScoringConfig.ID, t.ID,
			t.ScoringConfig.CorrectPoints, t.ScoringConfig.WrongPoints, t.ScoringConfig.BlankPoints,
		)
		if err != nil {
			return fmt.Errorf("insert scoring config: %w", err)
		}
	}

	if len(t.Questions) > 0 {
		if err := insertTestQuestions(ctx, tx, t.ID, t.Questions); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *TestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Test, error) {
	t := &domain.Test{}
	var cat, diff, status string
	var edulevel *string
	err := r.pool.QueryRow(ctx,
		`SELECT id, contributor_id, title, description, category, duration_minutes, difficulty,
		        status, education_level, created_at, published_at
		 FROM tests WHERE id = $1`, id,
	).Scan(&t.ID, &t.ContributorID, &t.Title, &t.Description, &cat, &t.DurationMinutes,
		&diff, &status, &edulevel, &t.CreatedAt, &t.PublishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find test: %w", err)
	}
	t.Category = domain.TestCategory(cat)
	t.Difficulty = domain.Difficulty(diff)
	t.Status = domain.TestStatus(status)
	if edulevel != nil {
		el := domain.EducationLevel(*edulevel)
		t.EducationLevel = &el
	}

	if err := r.loadQuestions(ctx, t); err != nil {
		return nil, err
	}
	if err := r.loadScoringConfig(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TestRepository) List(ctx context.Context, f repository.TestFilter) ([]*domain.Test, int, error) {
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

	if f.ContributorID != nil {
		conds = append(conds, fmt.Sprintf("contributor_id = $%d", i))
		args = append(args, *f.ContributorID)
		i++
	}
	if f.Category != nil {
		conds = append(conds, fmt.Sprintf("category = $%d", i))
		args = append(args, string(*f.Category))
		i++
	}
	if f.Difficulty != nil {
		conds = append(conds, fmt.Sprintf("difficulty = $%d", i))
		args = append(args, string(*f.Difficulty))
		i++
	}
	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", i))
		args = append(args, string(*f.Status))
		i++
	}
	if f.EducationLevel != nil {
		conds = append(conds, fmt.Sprintf("(education_level IS NULL OR education_level = $%d)", i))
		args = append(args, string(*f.EducationLevel))
		i++
	}

	where := buildWhere(conds)
	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM tests %s`, where), countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tests: %w", err)
	}

	args = append(args, f.Limit, offset)
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT id, contributor_id, title, description, category, duration_minutes,
		             difficulty, status, education_level, created_at, published_at
		             FROM tests %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			where, i, i+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list tests: %w", err)
	}
	defer rows.Close()

	var tests []*domain.Test
	for rows.Next() {
		t := &domain.Test{}
		var cat, diff, status string
		var edulevel *string
		if err := rows.Scan(&t.ID, &t.ContributorID, &t.Title, &t.Description, &cat,
			&t.DurationMinutes, &diff, &status, &edulevel, &t.CreatedAt, &t.PublishedAt); err != nil {
			return nil, 0, fmt.Errorf("scan test: %w", err)
		}
		t.Category = domain.TestCategory(cat)
		t.Difficulty = domain.Difficulty(diff)
		t.Status = domain.TestStatus(status)
		if edulevel != nil {
			el := domain.EducationLevel(*edulevel)
			t.EducationLevel = &el
		}
		tests = append(tests, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Compute per-test student status if caller is a student.
	if f.StudentID != nil {
		for _, t := range tests {
			var resultID *uuid.UUID
			_ = r.pool.QueryRow(ctx,
				`SELECT id FROM test_results WHERE student_id = $1 AND test_id = $2`,
				*f.StudentID, t.ID,
			).Scan(resultID) // may be pgx.ErrNoRows → nil
			var inProgress bool
			_ = r.pool.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM test_sessions WHERE student_id = $1 AND test_id = $2 AND status = 'in_progress')`,
				*f.StudentID, t.ID,
			).Scan(&inProgress)
			if resultID != nil {
				t.StudentStatus = "completed"
				t.ResultID = resultID
			} else if inProgress {
				t.StudentStatus = "in_progress"
			} else {
				t.StudentStatus = "not_started"
			}
		}
	}

	for _, t := range tests {
		if err := r.loadQuestions(ctx, t); err != nil {
			return nil, 0, err
		}
		if err := r.loadScoringConfig(ctx, t); err != nil {
			return nil, 0, err
		}
	}
	return tests, total, nil
}

func (r *TestRepository) Update(ctx context.Context, t *domain.Test) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE tests SET title=$1, description=$2, category=$3, duration_minutes=$4, difficulty=$5, education_level=$6
		 WHERE id=$7`,
		t.Title, t.Description, string(t.Category), t.DurationMinutes, string(t.Difficulty), t.EducationLevel, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update test: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *TestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tests WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete test: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *TestRepository) SetStatus(ctx context.Context, id uuid.UUID, status domain.TestStatus, publishedAt *time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE tests SET status=$1, published_at=$2 WHERE id=$3`,
		string(status), publishedAt, id,
	)
	if err != nil {
		return fmt.Errorf("set test status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *TestRepository) SetQuestions(ctx context.Context, testID uuid.UUID, questions []domain.TestQuestion) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM test_questions WHERE test_id = $1`, testID); err != nil {
		return fmt.Errorf("clear test questions: %w", err)
	}
	if len(questions) > 0 {
		if err := insertTestQuestions(ctx, tx, testID, questions); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *TestRepository) UpdateScoringConfig(ctx context.Context, sc *domain.ScoringConfig) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scoring_configs SET correct_points=$1, wrong_points=$2, blank_points=$3
		 WHERE test_id=$4`,
		sc.CorrectPoints, sc.WrongPoints, sc.BlankPoints, sc.TestID,
	)
	if err != nil {
		return fmt.Errorf("update scoring config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *TestRepository) QuestionCount(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM test_questions WHERE test_id = $1`, id,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count test questions: %w", err)
	}
	return count, nil
}

func (r *TestRepository) loadQuestions(ctx context.Context, t *domain.Test) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, test_id, question_id, order_index
		 FROM test_questions WHERE test_id = $1 ORDER BY order_index`, t.ID,
	)
	if err != nil {
		return fmt.Errorf("load test questions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var q domain.TestQuestion
		if err := rows.Scan(&q.ID, &q.TestID, &q.QuestionID, &q.OrderIndex); err != nil {
			return fmt.Errorf("scan test question: %w", err)
		}
		t.Questions = append(t.Questions, q)
	}
	return rows.Err()
}

func (r *TestRepository) loadScoringConfig(ctx context.Context, t *domain.Test) error {
	sc := &domain.ScoringConfig{TestID: t.ID}
	err := r.pool.QueryRow(ctx,
		`SELECT id, correct_points, wrong_points, blank_points
		 FROM scoring_configs WHERE test_id = $1`, t.ID,
	).Scan(&sc.ID, &sc.CorrectPoints, &sc.WrongPoints, &sc.BlankPoints)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("load scoring config: %w", err)
	}
	t.ScoringConfig = sc
	return nil
}

func insertTestQuestions(ctx context.Context, tx pgx.Tx, testID uuid.UUID, qs []domain.TestQuestion) error {
	for _, q := range qs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO test_questions (id, test_id, question_id, order_index) VALUES ($1,$2,$3,$4)`,
			q.ID, testID, q.QuestionID, q.OrderIndex,
		); err != nil {
			return fmt.Errorf("insert test question: %w", err)
		}
	}
	return nil
}

func buildWhere(conds []string) string {
	if len(conds) == 0 {
		return ""
	}
	w := "WHERE "
	for i, c := range conds {
		if i > 0 {
			w += " AND "
		}
		w += c
	}
	return w
}
