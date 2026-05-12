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
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users WHERE role = 'student'                        ) AS total_students,
			(SELECT COUNT(*) FROM users WHERE role = 'contributor' AND status = 'active') AS total_contributors,
			(SELECT COUNT(*) FROM tests                                               ) AS total_tests,
			(SELECT COUNT(*) FROM questions                                           ) AS total_questions,
			(SELECT COUNT(*) FROM users WHERE role = 'contributor' AND status = 'pending') AS pending_approvals
	`).Scan(
		&s.TotalStudents,
		&s.TotalContributors,
		&s.TotalTests,
		&s.TotalQuestions,
		&s.PendingApprovals,
	)
	if err != nil {
		return nil, fmt.Errorf("get platform stats: %w", err)
	}
	return &s, nil
}

func (r *AdminRepository) ListTestsWithAttempts(ctx context.Context, page, limit int) ([]*domain.TestWithAttempts, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tests`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tests: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id, t.contributor_id, t.title, t.category, t.difficulty, t.status,
			t.duration_minutes, t.created_at,
			COUNT(ts.id) AS attempt_count
		FROM tests t
		LEFT JOIN test_sessions ts ON ts.test_id = t.id
		GROUP BY t.id
		ORDER BY attempt_count DESC, t.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list tests with attempts: %w", err)
	}
	defer rows.Close()

	var out []*domain.TestWithAttempts
	for rows.Next() {
		var ta domain.TestWithAttempts
		var cat, diff, status string
		if err := rows.Scan(
			&ta.ID, &ta.ContributorID, &ta.Title, &cat, &diff, &status,
			&ta.DurationMinutes, &ta.CreatedAt, &ta.AttemptCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan test with attempts: %w", err)
		}
		ta.Category = domain.TestCategory(cat)
		ta.Difficulty = domain.Difficulty(diff)
		ta.Status = domain.TestStatus(status)
		out = append(out, &ta)
	}
	return out, total, rows.Err()
}
