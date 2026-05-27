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
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
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
		`SELECT tr.id, tr.session_id, tr.student_id, tr.test_id, t.title,
		        tr.total_score, tr.correct_count, tr.wrong_count, tr.blank_count, tr.completed_at
		 FROM test_results tr JOIN tests t ON t.id = tr.test_id
		 WHERE tr.session_id = $1`, sessionID,
	).Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID, &res.TestTitle,
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
		`SELECT tr.id, tr.session_id, tr.student_id, tr.test_id, t.title,
		        tr.total_score, tr.correct_count, tr.wrong_count, tr.blank_count, tr.completed_at
		 FROM test_results tr JOIN tests t ON t.id = tr.test_id
		 WHERE tr.student_id = $1 ORDER BY tr.completed_at DESC`, studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}
	defer rows.Close()

	var results []*domain.TestResult
	for rows.Next() {
		res := &domain.TestResult{}
		if err := rows.Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID, &res.TestTitle,
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
		`SELECT tr.id, tr.session_id, tr.student_id, tr.test_id, t.title,
		        tr.total_score, tr.correct_count, tr.wrong_count, tr.blank_count, tr.completed_at
		 FROM test_results tr JOIN tests t ON t.id = tr.test_id
		 WHERE tr.id = $1`, id,
	).Scan(&res.ID, &res.SessionID, &res.StudentID, &res.TestID, &res.TestTitle,
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
	// 1. Load base question info.
	rows, err := r.pool.Query(ctx,
		`SELECT q.id, q.question_type, tq.order_index, q.text, q.explanation, q.image_url, q.difficulty, tp.id, tp.name
		 FROM test_questions tq
		 JOIN questions q ON q.id = tq.question_id
		 JOIN topics tp ON tp.id = q.topic_id
		 WHERE tq.test_id = $1
		 ORDER BY tq.order_index`, testID,
	)
	if err != nil {
		return nil, fmt.Errorf("review query: %w", err)
	}
	defer rows.Close()

	var items []domain.ReviewItem
	questionIndex := map[uuid.UUID]int{}
	var questionIDs []uuid.UUID

	for rows.Next() {
		var item domain.ReviewItem
		var diff, qtype string
		if err := rows.Scan(&item.QuestionID, &qtype, &item.OrderIndex, &item.Text, &item.Explanation, &item.ImageURL, &diff, &item.TopicID, &item.TopicName); err != nil {
			return nil, fmt.Errorf("scan review base: %w", err)
		}
		item.Difficulty = domain.Difficulty(diff)
		item.QuestionType = domain.QuestionType(qtype)
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

	// 2. Load options for MCQ / multi_correct questions.
	optRows, err := r.pool.Query(ctx,
		`SELECT id, question_id, label, text, is_correct, image_url FROM question_options WHERE question_id = ANY($1) ORDER BY label`,
		questionIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load review options: %w", err)
	}
	defer optRows.Close()
	correctOptsByQ := map[uuid.UUID][]uuid.UUID{}
	for optRows.Next() {
		var opt domain.ReviewOption
		var qID uuid.UUID
		if err := optRows.Scan(&opt.ID, &qID, &opt.Label, &opt.Text, &opt.IsCorrect, &opt.ImageURL); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		if idx, ok := questionIndex[qID]; ok {
			items[idx].Options = append(items[idx].Options, opt)
		}
		if opt.IsCorrect {
			correctOptsByQ[qID] = append(correctOptsByQ[qID], opt.ID)
		}
	}
	if err := optRows.Err(); err != nil {
		return nil, err
	}

	// Set MCQ convenience fields.
	for i := range items {
		if items[i].QuestionType == domain.QuestionTypeMCQ {
			ids := correctOptsByQ[items[i].QuestionID]
			if len(ids) > 0 {
				items[i].CorrectOptionID = ids[0]
			}
		}
		if items[i].QuestionType == domain.QuestionTypeMultiCorrect {
			items[i].CorrectOptionIDs = correctOptsByQ[items[i].QuestionID]
		}
	}

	// 3. Load statements for true_false questions.
	stmtRows, err := r.pool.Query(ctx,
		`SELECT id, question_id, text, is_correct, position, image_url FROM question_statements WHERE question_id = ANY($1) ORDER BY position`,
		questionIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load review statements: %w", err)
	}
	defer stmtRows.Close()
	stmtsByQ := map[uuid.UUID][]domain.ReviewStatement{}
	for stmtRows.Next() {
		var rs domain.ReviewStatement
		var qID uuid.UUID
		var pos int
		if err := stmtRows.Scan(&rs.ID, &qID, &rs.Text, &rs.IsCorrect, &pos, &rs.ImageURL); err != nil {
			return nil, fmt.Errorf("scan statement: %w", err)
		}
		stmtsByQ[qID] = append(stmtsByQ[qID], rs)
	}
	if err := stmtRows.Err(); err != nil {
		return nil, err
	}

	// 4. Load student answers.
	ansRows, err := r.pool.Query(ctx,
		`SELECT question_id, selected_option_id, statement_id, boolean_answer FROM session_answers WHERE session_id = $1`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("load student answers: %w", err)
	}
	defer ansRows.Close()

	// Group by question.
	type studentAns struct {
		optionIDs   []uuid.UUID
		stmtAnswers map[uuid.UUID]bool
	}
	studentAnswers := map[uuid.UUID]*studentAns{}
	for ansRows.Next() {
		var qID uuid.UUID
		var optID, stmtID *uuid.UUID
		var boolAns *bool
		if err := ansRows.Scan(&qID, &optID, &stmtID, &boolAns); err != nil {
			return nil, fmt.Errorf("scan student answer: %w", err)
		}
		if studentAnswers[qID] == nil {
			studentAnswers[qID] = &studentAns{stmtAnswers: make(map[uuid.UUID]bool)}
		}
		if optID != nil {
			studentAnswers[qID].optionIDs = append(studentAnswers[qID].optionIDs, *optID)
		}
		if stmtID != nil && boolAns != nil {
			studentAnswers[qID].stmtAnswers[*stmtID] = *boolAns
		}
	}
	if err := ansRows.Err(); err != nil {
		return nil, err
	}

	// 5. Compute status and attach student answers to each item.
	for i := range items {
		q := &items[i]
		sa := studentAnswers[q.QuestionID]

		switch q.QuestionType {
		case domain.QuestionTypeMCQ:
			if sa == nil || len(sa.optionIDs) == 0 {
				q.Status = domain.AnswerStatusBlank
			} else {
				q.SelectedOptionID = &sa.optionIDs[0]
				if sa.optionIDs[0] == q.CorrectOptionID {
					q.Status = domain.AnswerStatusCorrect
				} else {
					q.Status = domain.AnswerStatusWrong
				}
			}

		case domain.QuestionTypeMultiCorrect:
			if sa == nil || len(sa.optionIDs) == 0 {
				q.Status = domain.AnswerStatusBlank
			} else {
				q.SelectedOptionIDs = sa.optionIDs
				if setsEqualUUID(sa.optionIDs, q.CorrectOptionIDs) {
					q.Status = domain.AnswerStatusCorrect
				} else {
					q.Status = domain.AnswerStatusWrong
				}
			}

		case domain.QuestionTypeTrueFalse:
			stmts := stmtsByQ[q.QuestionID]
			if sa == nil || len(sa.stmtAnswers) == 0 {
				q.Status = domain.AnswerStatusBlank
				for _, s := range stmts {
					q.Statements = append(q.Statements, domain.ReviewStatement{ID: s.ID, Text: s.Text, IsCorrect: s.IsCorrect})
				}
			} else {
				allCorrect := true
				for _, s := range stmts {
					ans, answered := sa.stmtAnswers[s.ID]
					var ansPtr *bool
					if answered {
						a := ans
						ansPtr = &a
						if ans != s.IsCorrect {
							allCorrect = false
						}
					} else {
						allCorrect = false
					}
					q.Statements = append(q.Statements, domain.ReviewStatement{ID: s.ID, Text: s.Text, IsCorrect: s.IsCorrect, StudentAnswer: ansPtr})
				}
				if allCorrect {
					q.Status = domain.AnswerStatusCorrect
				} else {
					q.Status = domain.AnswerStatusWrong
				}
			}
		}
	}

	return items, nil
}

func (r *ResultRepository) ListByTestID(ctx context.Context, f repository.ContributorResultFilter) ([]domain.ContributorResultEntry, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM test_results WHERE test_id = $1`, f.TestID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count results by test: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT tr.id, tr.session_id, tr.student_id, u.name, u.email,
		        tr.test_id, t.title, tr.total_score, tr.correct_count, tr.wrong_count, tr.blank_count, tr.completed_at,
		        tr.irt_theta
		 FROM test_results tr
		 JOIN users u ON u.id = tr.student_id
		 JOIN tests t ON t.id = tr.test_id
		 WHERE tr.test_id = $1
		 ORDER BY tr.completed_at DESC
		 LIMIT $2 OFFSET $3`,
		f.TestID, f.Limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list results by test: %w", err)
	}
	defer rows.Close()

	var out []domain.ContributorResultEntry
	for rows.Next() {
		var e domain.ContributorResultEntry
		if err := rows.Scan(&e.ID, &e.SessionID, &e.StudentID, &e.StudentName, &e.StudentEmail,
			&e.TestID, &e.TestTitle, &e.TotalScore, &e.CorrectCount, &e.WrongCount, &e.BlankCount, &e.CompletedAt,
			&e.IRTTheta); err != nil {
			return nil, 0, fmt.Errorf("scan result entry: %w", err)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *ResultRepository) GetTestAnalytics(ctx context.Context, testID uuid.UUID) (*domain.TestAnalytics, error) {
	a := &domain.TestAnalytics{}
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*),
		        AVG(total_score),
		        MAX(total_score),
		        MIN(total_score),
		        AVG(irt_theta)
		 FROM test_results WHERE test_id = $1`, testID,
	).Scan(&a.TotalAttempts, &a.AvgScore, &a.MaxScore, &a.MinScore, &a.AvgIRTTheta)
	if err != nil {
		return nil, fmt.Errorf("get test analytics: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT tp.id, tp.name,
		        COUNT(DISTINCT q.id)::int,
		        COALESCE(AVG(
		            CASE WHEN tq.question_id IS NOT NULL THEN
		                (SELECT COUNT(*) FROM session_answers sa
		                 JOIN question_options co ON co.question_id = sa.question_id AND co.is_correct = true AND co.id = sa.selected_option_id
		                 WHERE sa.session_id IN (SELECT id FROM test_sessions WHERE test_id = $1 AND status = 'submitted')
		                 AND sa.question_id = q.id)
		            END
		        ), 0)
		 FROM test_questions tq
		 JOIN questions q ON q.id = tq.question_id
		 JOIN topics tp ON tp.id = q.topic_id
		 WHERE tq.test_id = $1
		 GROUP BY tp.id, tp.name
		 ORDER BY tp.name`, testID,
	)
	if err != nil {
		return a, fmt.Errorf("per-topic analytics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ta domain.TopicAnalytics
		if err := rows.Scan(&ta.TopicID, &ta.TopicName, &ta.QuestionCount, &ta.AvgCorrectCount); err != nil {
			return a, fmt.Errorf("scan topic analytics: %w", err)
		}
		a.PerTopic = append(a.PerTopic, ta)
	}
	return a, rows.Err()
}

func setsEqualUUID(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[uuid.UUID]struct{}, len(b))
	for _, id := range b {
		set[id] = struct{}{}
	}
	for _, id := range a {
		if _, ok := set[id]; !ok {
			return false
		}
	}
	return true
}

func (r *ResultRepository) ExistsByStudentAndTest(ctx context.Context, studentID, testID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM test_results WHERE student_id = $1 AND test_id = $2)`,
		studentID, testID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check result exists: %w", err)
	}
	return exists, nil
}
