package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) GetStats(ctx context.Context) (*domain.PlatformStats, error) {
	var s domain.PlatformStats
	if err := r.pool.QueryRow(ctx,
		`SELECT
			(SELECT COUNT(*) FROM users WHERE role = 'student'),
			(SELECT COUNT(*) FROM users WHERE role = 'contributor'),
			(SELECT COUNT(*) FROM tests),
			(SELECT COUNT(*) FROM questions),
			(SELECT COUNT(*) FROM users WHERE role = 'contributor' AND status = 'pending'),
			(SELECT COUNT(*) FROM users WHERE role = 'student' AND created_at >= NOW() - INTERVAL '7 days'),
			(SELECT COUNT(*) FROM users WHERE role = 'contributor' AND created_at >= NOW() - INTERVAL '7 days'),
			(SELECT COUNT(*) FROM test_sessions WHERE status = 'submitted'),
			COALESCE((SELECT AVG(total_score) FROM test_results), 0),
			COALESCE((SELECT t.title FROM tests t JOIN test_sessions ts ON ts.test_id = t.id WHERE ts.status = 'submitted' GROUP BY t.title ORDER BY COUNT(*) DESC, t.title LIMIT 1), ''),
			COALESCE((SELECT COUNT(*) FROM test_sessions WHERE test_id = (SELECT test_id FROM test_sessions WHERE status = 'submitted' GROUP BY test_id ORDER BY COUNT(*) DESC LIMIT 1) AND status = 'submitted'), 0),
			COALESCE((SELECT COUNT(DISTINCT q.id) FROM questions q JOIN test_questions tq ON tq.question_id = q.id JOIN tests t ON t.id = tq.test_id WHERE t.category = 'tka_saintek'), 0),
			COALESCE((SELECT COUNT(DISTINCT q.id) FROM questions q JOIN test_questions tq ON tq.question_id = q.id JOIN tests t ON t.id = tq.test_id WHERE t.category = 'tka_soshum'), 0),
			COALESCE((SELECT COUNT(DISTINCT q.id) FROM questions q JOIN test_questions tq ON tq.question_id = q.id JOIN tests t ON t.id = tq.test_id WHERE t.category = 'smbt'), 0),
			(SELECT COUNT(*) FROM questions q WHERE EXISTS (SELECT 1 FROM test_questions tq WHERE tq.question_id = q.id)),
			(SELECT COUNT(*) FROM questions q WHERE NOT EXISTS (SELECT 1 FROM test_questions tq WHERE tq.question_id = q.id))`,
	).Scan(
		&s.TotalStudents, &s.TotalContributors, &s.TotalTests, &s.TotalQuestions, &s.PendingApprovals,
		&s.StudentsThisWeek, &s.ContributorsThisWeek, &s.TotalAttempts, &s.AvgScore,
		&s.TopTestTitle, &s.TopTestAttempts,
		&s.QuestionsTKASaintek, &s.QuestionsTKASoshum, &s.QuestionsSMBT,
		&s.QuestionsUsed, &s.QuestionsUnused,
	); err != nil {
		return nil, fmt.Errorf("get platform stats: %w", err)
	}
	return &s, nil
}

func (r *AdminRepository) ListTestsWithAttempts(ctx context.Context, page, limit int, educationLevel *domain.EducationLevel) ([]*domain.TestWithAttempts, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	args := []any{}
	where := ""
	if educationLevel != nil {
		where = fmt.Sprintf(" WHERE t.education_level = $%d OR t.education_level IS NULL", len(args)+1)
		args = append(args, string(*educationLevel))
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tests t`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tests: %w", err)
	}

	queryArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
			SELECT
				t.id, t.contributor_id, t.title, t.category, t.difficulty, t.status,
				t.duration_minutes, t.education_level, t.created_at,
				COUNT(ts.id) AS attempt_count
			FROM tests t
			LEFT JOIN test_sessions ts ON ts.test_id = t.id
			%s
			GROUP BY t.id
			ORDER BY attempt_count DESC, t.created_at DESC
			LIMIT $%d OFFSET $%d
		`, where, len(args)+1, len(args)+2), queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tests with attempts: %w", err)
	}
	defer rows.Close()

	var out []*domain.TestWithAttempts
	for rows.Next() {
		var ta domain.TestWithAttempts
		var cat, diff, status string
		var edulevel *string
		if err := rows.Scan(
			&ta.ID, &ta.ContributorID, &ta.Title, &cat, &diff, &status,
			&ta.DurationMinutes, &edulevel, &ta.CreatedAt, &ta.AttemptCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan test with attempts: %w", err)
		}
		ta.Category = domain.TestCategory(cat)
		ta.Difficulty = domain.Difficulty(diff)
		ta.Status = domain.TestStatus(status)
		if edulevel != nil {
			el := domain.EducationLevel(*edulevel)
			ta.EducationLevel = &el
		}
		out = append(out, &ta)
	}
	return out, total, rows.Err()
}
