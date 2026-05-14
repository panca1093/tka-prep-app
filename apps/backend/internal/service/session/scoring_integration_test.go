//go:build integration
// +build integration

package session

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	pgstore "github.com/yourorg/tkaprep/apps/backend/internal/repository/postgres"
)

// These tests run against a real Postgres instance.
// Set DATABASE_URL env var and use `go test -tags=integration`.
//
//	export DATABASE_URL="postgres://tkaprep:tkaprep@localhost:5432/tkaprep?sslmode=disable"
//	go test -tags=integration -v -run "TestScoringIntegration"

func TestScoringIntegration(t *testing.T) {
	ctx := context.Background()

	pool, err := pgstore.NewPool(ctx, "postgres://tkaprep:tkaprep@localhost:5432/tkaprep?sslmode=disable")
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to postgres (set DATABASE_URL or check docker): %v", err)
	}
	defer pool.Close()

	// ─── Setup: create a test with known scoring config ──────────────────
	contributorID := uuid.New()
	studentID := uuid.New()
	topicID := uuid.New()
	now := time.Now().UTC()

	// Insert topic
	_, err = pool.Exec(ctx, `INSERT INTO topics (id, name, created_at) VALUES ($1, $2, $3) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name`,
		topicID, "itg-topic-"+uuid.New().String()[:8], now)
	require.NoError(t, err)

	// Insert contributor (student role, we just need the user row)
	_, err = pool.Exec(ctx, `INSERT INTO users (id, name, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, 'c', $2, '$2a$10$dummy', 'contributor', 'active', $3, $3) ON CONFLICT (id) DO NOTHING`,
		contributorID, "ctest-"+uuid.New().String()[:8]+"@test.com", now)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `INSERT INTO users (id, name, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, 's', $2, '$2a$10$dummy', 'student', 'active', $3, $3) ON CONFLICT (id) DO NOTHING`,
		studentID, "stest-"+uuid.New().String()[:8]+"@test.com", now)
	require.NoError(t, err)

	// Insert test with scoring: +4 correct, 0 wrong, 0 blank
	testID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO tests (id, contributor_id, title, category, duration_minutes, difficulty, status, created_at)
		VALUES ($1, $2, 'Scoring Test', 'tka_saintek', 60, 'medium', 'published', $3)`, testID, contributorID, now)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `INSERT INTO scoring_configs (id, test_id, correct_points, wrong_points, blank_points)
		VALUES ($1, $2, 4, 0, 0)`, uuid.New(), testID)
	require.NoError(t, err)

	// Insert 3 MCQ questions with options
	qIDs := make([]uuid.UUID, 3)
	correctOptIDs := make([]uuid.UUID, 3)
	wrongOptIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		qID := uuid.New()
		qIDs[i] = qID
		correctOptIDs[i] = uuid.New()
		wrongOptIDs[i] = uuid.New()

		_, err = pool.Exec(ctx, `INSERT INTO questions (id, contributor_id, topic_id, question_type, text, difficulty, created_at, updated_at)
			VALUES ($1, $2, $3, 'mcq', 'Q', 'easy', $4, $4)`, qID, contributorID, topicID, now)
		require.NoError(t, err)

		// Insert 5 options, first is correct
		labels := []string{"A", "B", "C", "D", "E"}
		for j, l := range labels {
			optID := uuid.New()
			if j == 0 {
				correctOptIDs[i] = optID
			} else if j == 1 {
				wrongOptIDs[i] = optID
			}
			_, err = pool.Exec(ctx, `INSERT INTO question_options (id, question_id, label, text, is_correct)
				VALUES ($1, $2, $3, 'opt', $4)`, optID, qID, l, j == 0)
			require.NoError(t, err)
		}

		// Link question to test
		_, err = pool.Exec(ctx, `INSERT INTO test_questions (id, test_id, question_id, order_index)
			VALUES ($1, $2, $3, $4)`, uuid.New(), testID, qID, i)
		require.NoError(t, err)
	}

	// Create a session
	sessionID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO test_sessions (id, student_id, test_id, started_at, status, time_remaining_seconds)
		VALUES ($1, $2, $3, $4, 'in_progress', 3600)`, sessionID, studentID, testID, now)
	require.NoError(t, err)

	// Insert answers: 2 correct, 1 wrong → expected score = 2*4 + 1*0 = 8
	insertAnswer := func(sessID, qID, optID uuid.UUID) {
		_, err = pool.Exec(ctx, `INSERT INTO session_answers (id, session_id, question_id, selected_option_id, answered_at)
			VALUES ($1, $2, $3, $4, $5)`, uuid.New(), sessID, qID, optID, now)
		require.NoError(t, err)
	}
	insertAnswer(sessionID, qIDs[0], correctOptIDs[0]) // correct
	insertAnswer(sessionID, qIDs[1], correctOptIDs[1]) // correct
	insertAnswer(sessionID, qIDs[2], wrongOptIDs[2])   // wrong

	// ─── Submit via real service ──────────────────────────────────────
	sessRepo := pgstore.NewSessionRepository(pool)
	resultRepo := pgstore.NewResultRepository(pool)
	testRepo := pgstore.NewTestRepository(pool)
	svc := NewService(sessRepo, resultRepo, testRepo, pool)

	result, err := svc.Submit(ctx, sessionID, studentID)
	require.NoError(t, err)

	// ─── Verify scoring ───────────────────────────────────────────────
	require.Equal(t, 2, result.CorrectCount, "2 correct answers expected")
	require.Equal(t, 1, result.WrongCount, "1 wrong answer expected")
	require.Equal(t, 0, result.BlankCount, "0 blank answers expected")
	require.Equal(t, float64(8), result.TotalScore, "score = 2×4 + 1×0 = 8")

	// Verify session status updated
	var status string
	err = pool.QueryRow(ctx, `SELECT status FROM test_sessions WHERE id = $1`, sessionID).Scan(&status)
	require.NoError(t, err)
	require.Equal(t, string(domain.SessionStatusSubmitted), status)

	// ─── Cleanup ──────────────────────────────────────────────────────
	pool.Exec(ctx, `DELETE FROM test_results WHERE session_id = $1`, sessionID)
	pool.Exec(ctx, `DELETE FROM session_answers WHERE session_id = $1`, sessionID)
	pool.Exec(ctx, `DELETE FROM test_sessions WHERE id = $1`, sessionID)
	pool.Exec(ctx, `DELETE FROM test_questions WHERE test_id = $1`, testID)
	pool.Exec(ctx, `DELETE FROM question_options WHERE question_id = ANY($1)`, qIDs)
	pool.Exec(ctx, `DELETE FROM questions WHERE id = ANY($1)`, qIDs)
	pool.Exec(ctx, `DELETE FROM scoring_configs WHERE test_id = $1`, testID)
	pool.Exec(ctx, `DELETE FROM tests WHERE id = $1`, testID)
	pool.Exec(ctx, `DELETE FROM topics WHERE id = $1`, topicID)
	pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1, $2)`, contributorID, studentID)
}
