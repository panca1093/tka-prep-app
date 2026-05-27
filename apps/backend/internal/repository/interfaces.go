package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
)

// UserAdminFilter holds optional filters for the admin user listing.
type UserAdminFilter struct {
	Search string
	Role   *domain.Role
	Status *domain.Status
	Page   int
	Limit  int
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// List returns a filtered, paginated list of users for admin views.
	List(ctx context.Context, f UserAdminFilter) ([]*domain.User, int, error)
	// UpdateStatus sets the account status for a user.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error
	// UpdateEducationLevel sets the education level for a user (students only).
	UpdateEducationLevel(ctx context.Context, id uuid.UUID, level *domain.EducationLevel) error
// UpdateProfile sets optional profile fields (gender, phone, avatar).
UpdateProfile(ctx context.Context, id uuid.UUID, gender *domain.Gender, phone *string, avatarURL *string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	DeleteByTokenHash(ctx context.Context, hash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type TopicRepository interface {
	List(ctx context.Context) ([]*domain.Topic, error)
	Create(ctx context.Context, t *domain.Topic) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
	Update(ctx context.Context, t *domain.Topic) error
	Delete(ctx context.Context, id uuid.UUID) error
	// IsUsedByQuestions reports whether any question references the topic.
	IsUsedByQuestions(ctx context.Context, id uuid.UUID) (bool, error)
}

// QuestionFilter holds optional filters for listing questions.
type QuestionFilter struct {
	Search         string
	TopicID        *uuid.UUID
	Difficulty     *domain.Difficulty
	QuestionType   *domain.QuestionType
	EducationLevel *string
	CallerID       uuid.UUID // if set, usage stats populated when CallerID == question.contributor_id
	Page           int
	Limit          int
}

type QuestionRepository interface {
	List(ctx context.Context, f QuestionFilter) ([]*domain.Question, int, error)
	Create(ctx context.Context, q *domain.Question) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	Update(ctx context.Context, q *domain.Question) error
	Delete(ctx context.Context, id uuid.UUID) error
	// IsUsedInPublishedTest reports whether the question is referenced by any published test.
	IsUsedInPublishedTest(ctx context.Context, id uuid.UUID) (bool, error)
	// ListUsageStats returns questions owned by contributor ranked by total test attachment count.
	ListUsageStats(ctx context.Context, contributorID uuid.UUID, limit int) ([]domain.QuestionUsageEntry, error)
}

// TestFilter holds optional filters for listing tests.
type TestFilter struct {
	ContributorID  *uuid.UUID
	CategoryID     *uuid.UUID
	Difficulty     *domain.Difficulty
	Status         *domain.TestStatus
	EducationLevel *domain.EducationLevel
	StudentID      *uuid.UUID // if set, compute student_status + result_id per test
	Page           int
	Limit          int
}

type TestRepository interface {
	// Create inserts the test and its scoring config atomically.
	Create(ctx context.Context, t *domain.Test) error
	// FindByID returns the test with questions and scoring config populated.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Test, error)
	// List returns test summaries (no questions/scoring loaded).
	List(ctx context.Context, f TestFilter) ([]*domain.Test, int, error)
	Update(ctx context.Context, t *domain.Test) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetStatus(ctx context.Context, id uuid.UUID, status domain.TestStatus, publishedAt *time.Time) error
	// SetQuestions replaces the full question list for a test.
	SetQuestions(ctx context.Context, testID uuid.UUID, questions []domain.TestQuestion) error
	// UpdateScoringConfig updates correct/wrong/blank points for a test.
	UpdateScoringConfig(ctx context.Context, sc *domain.ScoringConfig) error
	// QuestionCount returns the number of questions in a test.
	QuestionCount(ctx context.Context, id uuid.UUID) (int, error)
}

type SessionRepository interface {
	Create(ctx context.Context, s *domain.TestSession) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.TestSession, error)
	FindActiveByStudentAndTest(ctx context.Context, studentID, testID uuid.UUID) (*domain.TestSession, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SessionStatus, submittedAt *time.Time, timeRemaining int) error
	SaveMCQAnswer(ctx context.Context, sessionID, questionID uuid.UUID, optionID *uuid.UUID, now time.Time) error
	SavePGKAnswers(ctx context.Context, sessionID, questionID uuid.UUID, optionIDs []uuid.UUID, now time.Time) error
	SaveBSAnswers(ctx context.Context, sessionID, questionID uuid.UUID, stmtAnswers []domain.StatementAnswerInput, now time.Time) error
	UpsertFlag(ctx context.Context, sessionID, questionID uuid.UUID, flagged bool) error
}

type LeaderboardRepository interface {
	// List returns up to 100 entries ranked by aggregate total_score for the given scope.
	List(ctx context.Context, scope domain.LeaderboardScope) ([]domain.LeaderboardEntry, error)
	// GetMyRank returns the caller's rank and aggregate stats in the given scope.
	// Returns apierr.ErrNotFound if the student has no results in the scope.
	GetMyRank(ctx context.Context, studentID uuid.UUID, scope domain.LeaderboardScope) (*domain.LeaderboardEntry, error)
	// FindByTest returns ranked results for a specific test, up to limit.
	FindByTest(ctx context.Context, testID uuid.UUID, limit int) ([]domain.TestLeaderboardEntry, error)
}

type AdminRepository interface {
	// GetStats returns aggregate platform statistics.
	GetStats(ctx context.Context) (*domain.PlatformStats, error)
	// ListTestsWithAttempts returns tests with their attempt counts, ordered by attempt count desc.
	ListTestsWithAttempts(ctx context.Context, page, limit int, educationLevel *domain.EducationLevel, categoryID *uuid.UUID) ([]*domain.TestWithAttempts, int, error)
}

type CategoryRepository interface {
	List(ctx context.Context) ([]*domain.Category, error)
	Create(ctx context.Context, name string) (*domain.Category, error)
	Update(ctx context.Context, id uuid.UUID, name string) (*domain.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	TestCount(ctx context.Context, id uuid.UUID) (int, error)
}

type ResultRepository interface {
	// Create inserts a new test result.
	Create(ctx context.Context, r *domain.TestResult) error
	// FindBySessionID returns the result for a given session.
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) (*domain.TestResult, error)
	// ListByStudent returns all results for a student, newest first.
	ListByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.TestResult, error)
	// ListAll returns all results across all students, newest first (admin use).
	ListAll(ctx context.Context) ([]*domain.TestResult, error)
	// FindByID returns a single result by its own primary key.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.TestResult, error)
	// FindDetailByID returns the result enriched with per-topic breakdown.
	FindDetailByID(ctx context.Context, id uuid.UUID) (*domain.ResultDetail, error)
	// GetReview returns per-question review items for a completed session.
	GetReview(ctx context.Context, sessionID, testID uuid.UUID) ([]domain.ReviewItem, error)
	// ListByTestID returns paginated results for a test, with student name/email joined.
	ListByTestID(ctx context.Context, f ContributorResultFilter) ([]domain.ContributorResultEntry, int, error)
	// GetTestAnalytics returns aggregate stats + per-topic breakdown for a test.
	GetTestAnalytics(ctx context.Context, testID uuid.UUID) (*domain.TestAnalytics, error)
	// ExistsByStudentAndTest reports whether the student has a completed result for the test.
	ExistsByStudentAndTest(ctx context.Context, studentID, testID uuid.UUID) (bool, error)
}

// ContributorResultFilter holds pagination for listing results by test.
type ContributorResultFilter struct {
	TestID uuid.UUID
	Page   int
	Limit  int
}
