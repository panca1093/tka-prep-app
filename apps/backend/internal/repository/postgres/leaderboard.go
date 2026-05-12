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

type LeaderboardRepository struct {
	pool *pgxpool.Pool
}

func NewLeaderboardRepository(pool *pgxpool.Pool) *LeaderboardRepository {
	return &LeaderboardRepository{pool: pool}
}

func (r *LeaderboardRepository) List(ctx context.Context, scope domain.LeaderboardScope) ([]domain.LeaderboardEntry, error) {
	q, args := leaderboardQuery(scope, uuid.Nil, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("leaderboard list: %w", err)
	}
	defer rows.Close()

	var entries []domain.LeaderboardEntry
	for rows.Next() {
		var e domain.LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.StudentID, &e.StudentName, &e.TotalScore, &e.TestCount); err != nil {
			return nil, fmt.Errorf("scan leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *LeaderboardRepository) GetMyRank(ctx context.Context, studentID uuid.UUID, scope domain.LeaderboardScope) (*domain.LeaderboardEntry, error) {
	q, args := leaderboardQuery(scope, studentID, true)
	e := &domain.LeaderboardEntry{}
	err := r.pool.QueryRow(ctx, q, args...).Scan(&e.Rank, &e.StudentID, &e.StudentName, &e.TotalScore, &e.TestCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("leaderboard my rank: %w", err)
	}
	return e, nil
}

// leaderboardQuery builds the SQL and args for both the list and my-rank queries.
// When forStudent is true, a CTE wraps the ranking and filters to the given studentID.
func leaderboardQuery(scope domain.LeaderboardScope, studentID uuid.UUID, forStudent bool) (string, []interface{}) {
	// Build the WHERE clause and any extra JOIN for scope.
	// Arg $1 is reserved for studentID when forStudent=true; otherwise no positional arg yet.
	// We use a different arg index baseline depending on forStudent.
	var (
		join  string
		where string
		args  []interface{}
	)

	if forStudent {
		args = append(args, studentID) // $1
	}

	nextArg := len(args) + 1

	switch scope {
	case domain.ScopeTKA:
		join = "JOIN tests t ON t.id = tr.test_id"
		where = fmt.Sprintf("WHERE t.category IN ($%d, $%d)", nextArg, nextArg+1)
		args = append(args, domain.CategoryTKASaintek, domain.CategoryTKASoshum)
	case domain.ScopeSMBT:
		join = "JOIN tests t ON t.id = tr.test_id"
		where = fmt.Sprintf("WHERE t.category = $%d", nextArg)
		args = append(args, domain.CategorySMBT)
	case domain.ScopeWeek:
		where = fmt.Sprintf("WHERE tr.completed_at >= NOW() - INTERVAL '7 days'")
	}

	baseSQL := fmt.Sprintf(`
		SELECT
		    RANK() OVER (ORDER BY SUM(tr.total_score) DESC) AS rank,
		    tr.student_id,
		    u.name AS student_name,
		    SUM(tr.total_score)            AS total_score,
		    COUNT(DISTINCT tr.test_id)::int AS test_count
		FROM test_results tr
		JOIN users u ON u.id = tr.student_id
		%s
		%s
		GROUP BY tr.student_id, u.name`, join, where)

	if forStudent {
		q := fmt.Sprintf(`WITH ranked AS (%s) SELECT rank, student_id, student_name, total_score, test_count FROM ranked WHERE student_id = $1`, baseSQL)
		return q, args
	}

	q := baseSQL + "\nORDER BY rank\nLIMIT 100"
	return q, args
}
