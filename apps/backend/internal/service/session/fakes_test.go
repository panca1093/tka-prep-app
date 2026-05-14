package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

// ─── call-tracking types ──────────────────────────────────────────────────────

type statusUpdate struct {
	id            uuid.UUID
	status        domain.SessionStatus
	timeRemaining int
}

type flagUpdate struct {
	sessionID  uuid.UUID
	questionID uuid.UUID
	flagged    bool
}

// ─── fakeSessionRepo ─────────────────────────────────────────────────────────

type fakeSessionRepo struct {
	byID   map[uuid.UUID]*domain.TestSession
	active map[string]*domain.TestSession // key: studentID|testID

	created  []*domain.TestSession
	statuses []statusUpdate
	answers  []*domain.SessionAnswer
	flags    []flagUpdate
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{
		byID:   make(map[uuid.UUID]*domain.TestSession),
		active: make(map[string]*domain.TestSession),
	}
}

func activeKey(studentID, testID uuid.UUID) string {
	return studentID.String() + "|" + testID.String()
}

// seed stores a session in the fake without recording it as a "created" call.
func (r *fakeSessionRepo) seed(s *domain.TestSession) {
	cp := *s
	if s.Answers != nil {
		cp.Answers = make([]domain.SessionAnswer, len(s.Answers))
		copy(cp.Answers, s.Answers)
	}
	r.byID[cp.ID] = &cp
	if cp.Status == domain.SessionStatusInProgress {
		r.active[activeKey(cp.StudentID, cp.TestID)] = &cp
	}
}

func (r *fakeSessionRepo) Create(_ context.Context, s *domain.TestSession) error {
	cp := *s
	r.byID[cp.ID] = &cp
	r.created = append(r.created, &cp)
	if cp.Status == domain.SessionStatusInProgress {
		r.active[activeKey(cp.StudentID, cp.TestID)] = &cp
	}
	return nil
}

func (r *fakeSessionRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.TestSession, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	cp := *s
	cp.Answers = make([]domain.SessionAnswer, len(s.Answers))
	copy(cp.Answers, s.Answers)
	return &cp, nil
}

func (r *fakeSessionRepo) FindActiveByStudentAndTest(_ context.Context, studentID, testID uuid.UUID) (*domain.TestSession, error) {
	s, ok := r.active[activeKey(studentID, testID)]
	if !ok {
		return nil, nil
	}
	cp := *s
	cp.Answers = make([]domain.SessionAnswer, len(s.Answers))
	copy(cp.Answers, s.Answers)
	return &cp, nil
}

func (r *fakeSessionRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.SessionStatus, _ *time.Time, timeRemaining int) error {
	r.statuses = append(r.statuses, statusUpdate{id: id, status: status, timeRemaining: timeRemaining})
	if s, ok := r.byID[id]; ok {
		s.Status = status
		s.TimeRemainingSeconds = timeRemaining
		if status != domain.SessionStatusInProgress {
			delete(r.active, activeKey(s.StudentID, s.TestID))
		}
	}
	return nil
}

func (r *fakeSessionRepo) UpsertAnswer(_ context.Context, a *domain.SessionAnswer) error {
	cp := *a
	r.answers = append(r.answers, &cp)
	if sess, ok := r.byID[a.SessionID]; ok {
		for i, existing := range sess.Answers {
			if existing.QuestionID == a.QuestionID {
				sess.Answers[i] = cp
				return nil
			}
		}
		sess.Answers = append(sess.Answers, cp)
	}
	return nil
}

func (r *fakeSessionRepo) SaveMCQAnswer(_ context.Context, sessionID, questionID uuid.UUID, optionID *uuid.UUID, _ time.Time) error {
	optID := uuid.Nil
	if optionID != nil {
		optID = *optionID
	}
	r.answers = append(r.answers, &domain.SessionAnswer{
		ID: uuid.New(), SessionID: sessionID, QuestionID: questionID,
		SelectedOptionID: &optID, AnsweredAt: time.Now().UTC(),
	})
	return nil
}

func (r *fakeSessionRepo) SavePGKAnswers(_ context.Context, sessionID, questionID uuid.UUID, optionIDs []uuid.UUID, _ time.Time) error {
	for _, oid := range optionIDs {
		r.answers = append(r.answers, &domain.SessionAnswer{
			ID: uuid.New(), SessionID: sessionID, QuestionID: questionID,
			SelectedOptionID: &oid, AnsweredAt: time.Now().UTC(),
		})
	}
	return nil
}

func (r *fakeSessionRepo) SaveBSAnswers(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ []domain.StatementAnswerInput, _ time.Time) error {
	return nil
}

func (r *fakeSessionRepo) UpsertFlag(_ context.Context, sessionID, questionID uuid.UUID, flagged bool) error {
	r.flags = append(r.flags, flagUpdate{sessionID: sessionID, questionID: questionID, flagged: flagged})
	return nil
}

func (r *fakeSessionRepo) LoadAnswers(_ context.Context, s *domain.TestSession) error {
	if stored, ok := r.byID[s.ID]; ok {
		s.Answers = make([]domain.SessionAnswer, len(stored.Answers))
		copy(s.Answers, stored.Answers)
	}
	return nil
}

// ─── fakeTestRepo ─────────────────────────────────────────────────────────────

type fakeTestRepo struct {
	tests map[uuid.UUID]*domain.Test
}

func newFakeTestRepo(tests ...*domain.Test) *fakeTestRepo {
	r := &fakeTestRepo{tests: make(map[uuid.UUID]*domain.Test)}
	for _, t := range tests {
		r.tests[t.ID] = t
	}
	return r
}

func (r *fakeTestRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Test, error) {
	t, ok := r.tests[id]
	if !ok {
		return nil, apierr.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *fakeTestRepo) Create(_ context.Context, _ *domain.Test) error { return nil }
func (r *fakeTestRepo) List(_ context.Context, _ repository.TestFilter) ([]*domain.Test, int, error) {
	return nil, 0, nil
}
func (r *fakeTestRepo) Update(_ context.Context, _ *domain.Test) error { return nil }
func (r *fakeTestRepo) Delete(_ context.Context, _ uuid.UUID) error     { return nil }
func (r *fakeTestRepo) SetStatus(_ context.Context, _ uuid.UUID, _ domain.TestStatus, _ *time.Time) error {
	return nil
}
func (r *fakeTestRepo) SetQuestions(_ context.Context, _ uuid.UUID, _ []domain.TestQuestion) error {
	return nil
}
func (r *fakeTestRepo) UpdateScoringConfig(_ context.Context, _ *domain.ScoringConfig) error {
	return nil
}
func (r *fakeTestRepo) QuestionCount(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }

// ─── fakeResultRepo ───────────────────────────────────────────────────────────

type fakeResultRepo struct{}

func (r *fakeResultRepo) Create(_ context.Context, _ *domain.TestResult) error { return nil }
func (r *fakeResultRepo) FindBySessionID(_ context.Context, _ uuid.UUID) (*domain.TestResult, error) {
	return nil, apierr.ErrNotFound
}
func (r *fakeResultRepo) ListByStudent(_ context.Context, _ uuid.UUID) ([]*domain.TestResult, error) {
	return nil, nil
}
func (r *fakeResultRepo) ListAll(_ context.Context) ([]*domain.TestResult, error) { return nil, nil }
func (r *fakeResultRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.TestResult, error) {
	return nil, apierr.ErrNotFound
}
func (r *fakeResultRepo) FindDetailByID(_ context.Context, _ uuid.UUID) (*domain.ResultDetail, error) {
	return nil, apierr.ErrNotFound
}
func (r *fakeResultRepo) GetReview(_ context.Context, _, _ uuid.UUID) ([]domain.ReviewItem, error) {
	return nil, nil
}

// ─── builder helpers ──────────────────────────────────────────────────────────

func makePublishedTest(id uuid.UUID, durationMinutes int) *domain.Test {
	sc := &domain.ScoringConfig{
		ID:            uuid.New(),
		TestID:        id,
		CorrectPoints: 4.0,
		WrongPoints:   -1.0,
		BlankPoints:   0.0,
	}
	return &domain.Test{
		ID:              id,
		ContributorID:   uuid.New(),
		Title:           "Sample Test",
		Category:        domain.CategoryTKASaintek,
		DurationMinutes: durationMinutes,
		Difficulty:      domain.DifficultyMedium,
		Status:          domain.TestStatusPublished,
		CreatedAt:       time.Now().UTC(),
		ScoringConfig:   sc,
	}
}

func makeDraftTest(id uuid.UUID) *domain.Test {
	t := makePublishedTest(id, 60)
	t.Status = domain.TestStatusDraft
	return t
}

// makeInProgressSession returns an in-progress session that started startedAgo ago.
func makeInProgressSession(id, studentID, testID uuid.UUID, startedAgo time.Duration) *domain.TestSession {
	return &domain.TestSession{
		ID:        id,
		StudentID: studentID,
		TestID:    testID,
		StartedAt: time.Now().Add(-startedAgo),
		Status:    domain.SessionStatusInProgress,
	}
}
