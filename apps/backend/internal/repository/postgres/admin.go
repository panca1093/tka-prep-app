package postgres

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
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
			(SELECT COUNT(*) FROM users WHERE role = 'contributor' AND created_at >= NOW() - INTERVAL '7 days')`,
	).Scan(
		&s.TotalStudents, &s.TotalContributors, &s.TotalTests, &s.TotalQuestions, &s.PendingApprovals,
		&s.StudentsThisWeek, &s.ContributorsThisWeek,
	); err != nil {
		return nil, fmt.Errorf("get platform stats: %w", err)
	}

	// --- attempts_by_topic ---
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.name AS topic,
			COALESCE(ts.education_level::text, '') AS edlevel,
			COUNT(DISTINCT tsess.id) AS attempt_count
		FROM test_sessions tsess
		JOIN tests ts ON ts.id = tsess.test_id
		JOIN test_questions tq ON tq.test_id = ts.id
		JOIN questions q ON q.id = tq.question_id
		JOIN topics t ON t.id = q.topic_id
		WHERE tsess.status = 'submitted'
		GROUP BY t.name, ts.education_level
		ORDER BY t.name, ts.education_level`)
	if err != nil {
		return nil, fmt.Errorf("get attempts by topic: %w", err)
	}

	type row struct {
		topic   string
		edlevel string
		count   int
	}
	raw := map[string]map[string]int{}
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.topic, &r.edlevel, &r.count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan attempts by topic: %w", err)
		}
		if raw[r.topic] == nil {
			raw[r.topic] = map[string]int{}
		}
		raw[r.topic][r.edlevel] = r.count
	}
	rows.Close()

	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s.AttemptsByTopic = append(s.AttemptsByTopic, domain.AttemptsByTopic{
			Topic: k,
			SD:    raw[k]["sd"],
			SMP:   raw[k]["smp"],
			SMA:   raw[k]["sma"],
			SMK:   raw[k]["smk"],
		})
	}
	if s.AttemptsByTopic == nil {
		s.AttemptsByTopic = []domain.AttemptsByTopic{}
	}

	// --- topic_performance ---
	rows2, err := r.pool.Query(ctx, `
		SELECT
			t.name AS topic,
			COALESCE(ROUND(AVG(tr.total_score)::numeric, 1), 0) AS avg_score,
			COALESCE(MAX(tr.total_score), 0) AS best_score
		FROM test_results tr
		JOIN test_sessions tsess ON tsess.id = tr.session_id
		JOIN tests ts ON ts.id = tr.test_id
		JOIN test_questions tq ON tq.test_id = ts.id
		JOIN questions q ON q.id = tq.question_id
		JOIN topics t ON t.id = q.topic_id
		GROUP BY t.name
		ORDER BY t.name`)
	if err != nil {
		return nil, fmt.Errorf("get topic performance: %w", err)
	}
	for rows2.Next() {
		var tp domain.TopicPerformance
		if err := rows2.Scan(&tp.Topic, &tp.AvgScore, &tp.BestScore); err != nil {
			rows2.Close()
			return nil, fmt.Errorf("scan topic performance: %w", err)
		}
		s.TopicPerformance = append(s.TopicPerformance, tp)
	}
	rows2.Close()
	if s.TopicPerformance == nil {
		s.TopicPerformance = []domain.TopicPerformance{}
	}

	// --- question_counts ---
	rows3, err := r.pool.Query(ctx, `
		SELECT
			q.education_level::text,
			t.name AS topic,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE EXISTS (
				SELECT 1 FROM test_questions tq WHERE tq.question_id = q.id
			)) AS used,
			COUNT(*) FILTER (WHERE NOT EXISTS (
				SELECT 1 FROM test_questions tq WHERE tq.question_id = q.id
			)) AS unused
		FROM questions q
		JOIN topics t ON t.id = q.topic_id
		WHERE q.education_level IS NOT NULL
		GROUP BY q.education_level, t.name
		ORDER BY q.education_level, t.name`)
	if err != nil {
		return nil, fmt.Errorf("get question counts: %w", err)
	}

	type qcRow struct {
		edlevel string
		topic   string
		total   int
		used    int
		unused  int
	}
	qcRaw := map[string][]domain.TopicQuestionCount{}
	for rows3.Next() {
		var qr qcRow
		if err := rows3.Scan(&qr.edlevel, &qr.topic, &qr.total, &qr.used, &qr.unused); err != nil {
			rows3.Close()
			return nil, fmt.Errorf("scan question counts: %w", err)
		}
		qcRaw[qr.edlevel] = append(qcRaw[qr.edlevel], domain.TopicQuestionCount{
			Topic:  qr.topic,
			Total:  qr.total,
			Used:   qr.used,
			Unused: qr.unused,
		})
	}
	rows3.Close()

	edOrder := []string{"sd", "smp", "sma", "smk"}
	for _, ed := range edOrder {
		if topics, ok := qcRaw[ed]; ok {
			s.QuestionCounts = append(s.QuestionCounts, domain.QuestionCountGroup{
				EducationLevel: ed,
				Topics:         topics,
			})
		}
	}
	if s.QuestionCounts == nil {
		s.QuestionCounts = []domain.QuestionCountGroup{}
	}

	// --- contributor_productivity ---
	rows4, err := r.pool.Query(ctx, `
		SELECT
			u.name,
			u.email,
			COUNT(DISTINCT q.id)::int AS question_count,
			COUNT(DISTINCT t.id)::int AS test_count
		FROM users u
		LEFT JOIN questions q ON q.contributor_id = u.id
		LEFT JOIN tests t ON t.contributor_id = u.id AND t.status = 'published'
		WHERE u.role = 'contributor'
		GROUP BY u.id, u.name, u.email
		HAVING COUNT(DISTINCT q.id) + COUNT(DISTINCT t.id) > 0
		ORDER BY (COUNT(DISTINCT q.id) + COUNT(DISTINCT t.id)) DESC, u.name`)
	if err != nil {
		return nil, fmt.Errorf("get contributor productivity: %w", err)
	}
	for rows4.Next() {
		var cp domain.ContributorProductivity
		if err := rows4.Scan(&cp.Name, &cp.Email, &cp.QuestionCount, &cp.TestCount); err != nil {
			rows4.Close()
			return nil, fmt.Errorf("scan contributor productivity: %w", err)
		}
		cp.OutputScore = cp.QuestionCount + cp.TestCount
		s.ContributorProductivity = append(s.ContributorProductivity, cp)
	}
	rows4.Close()
	if s.ContributorProductivity == nil {
		s.ContributorProductivity = []domain.ContributorProductivity{}
	}

	// --- test_completion ---
	rows5, err := r.pool.Query(ctx, `
		SELECT
			t.title,
			COUNT(ts.id)::int AS started,
			COUNT(*) FILTER (WHERE ts.status = 'submitted')::int AS submitted,
			COUNT(*) FILTER (WHERE ts.status = 'expired')::int AS expired
		FROM tests t
		JOIN test_sessions ts ON ts.test_id = t.id
		GROUP BY t.id, t.title
		HAVING COUNT(ts.id) > 0
		ORDER BY COUNT(*) FILTER (WHERE ts.status = 'submitted') DESC, t.title`)
	if err != nil {
		return nil, fmt.Errorf("get test completion: %w", err)
	}
	for rows5.Next() {
		var tc domain.TestCompletion
		if err := rows5.Scan(&tc.TestTitle, &tc.Started, &tc.Submitted, &tc.Expired); err != nil {
			rows5.Close()
			return nil, fmt.Errorf("scan test completion: %w", err)
		}
		terminal := tc.Submitted + tc.Expired
		if terminal > 0 {
			tc.CompletionPct = float64(tc.Submitted) / float64(terminal) * 100
		}
		s.TestCompletion = append(s.TestCompletion, tc)
	}
	rows5.Close()
	if s.TestCompletion == nil {
		s.TestCompletion = []domain.TestCompletion{}
	}

	// --- activity_feed ---
	rows6, err := r.pool.Query(ctx, `
		(SELECT
			'registration' AS event_type,
			u.name AS actor_name,
			'' AS detail,
			u.created_at AT TIME ZONE 'UTC' AS timestamp
		FROM users u
		WHERE u.role = 'student' AND u.created_at >= NOW() - INTERVAL '7 days')
		UNION ALL
		(SELECT
			'publication',
			c.name,
			t.title,
			t.published_at AT TIME ZONE 'UTC'
		FROM tests t
		JOIN users c ON c.id = t.contributor_id
		WHERE t.status = 'published' AND t.published_at >= NOW() - INTERVAL '7 days')
		UNION ALL
		(SELECT
			'submission',
			s.name,
			t.title || ' — Skor ' || ROUND(tr.total_score)::text,
			ts.submitted_at AT TIME ZONE 'UTC'
		FROM test_sessions ts
		JOIN users s ON s.id = ts.student_id
		JOIN tests t ON t.id = ts.test_id
		JOIN test_results tr ON tr.session_id = ts.id
		WHERE ts.status = 'submitted' AND ts.submitted_at >= NOW() - INTERVAL '7 days')
		ORDER BY timestamp DESC
		LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("get activity feed: %w", err)
	}
	for rows6.Next() {
		var af domain.ActivityFeedEntry
		if err := rows6.Scan(&af.EventType, &af.ActorName, &af.Detail, &af.Timestamp); err != nil {
			rows6.Close()
			return nil, fmt.Errorf("scan activity feed: %w", err)
		}
		s.ActivityFeed = append(s.ActivityFeed, af)
	}
	rows6.Close()
	if s.ActivityFeed == nil {
		s.ActivityFeed = []domain.ActivityFeedEntry{}
	}

	return &s, nil
}

func (r *AdminRepository) ListTestsWithAttempts(ctx context.Context, page, limit int, educationLevel *domain.EducationLevel, categoryID *uuid.UUID) ([]*domain.TestWithAttempts, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	args := []any{}
	conds := []string{}
	if educationLevel != nil {
		conds = append(conds, fmt.Sprintf("(t.education_level = $%d OR t.education_level IS NULL)", len(args)+1))
		args = append(args, string(*educationLevel))
	}
	if categoryID != nil {
		conds = append(conds, fmt.Sprintf("t.category_id = $%d", len(args)+1))
		args = append(args, *categoryID)
	}
	where := buildWhere(conds)

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tests t`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tests: %w", err)
	}

	queryArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
			SELECT
				t.id, t.contributor_id, t.title, COALESCE(c.name,''), t.difficulty, t.status,
				t.duration_minutes, t.education_level, t.created_at,
				COUNT(ts.id) AS attempt_count
			FROM tests t
			JOIN categories c ON c.id = t.category_id
			LEFT JOIN test_sessions ts ON ts.test_id = t.id
			%s
			GROUP BY t.id, c.name
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
		ta.CategoryName = cat
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
