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
			(SELECT COUNT(*) FROM users WHERE role = 'contributor' AND status = 'pending')`,
	).Scan(&s.TotalStudents, &s.TotalContributors, &s.TotalTests, &s.TotalQuestions, &s.PendingApprovals); err != nil {
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
