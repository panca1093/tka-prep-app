package domain

import (
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusSubmitted  SessionStatus = "submitted"
	SessionStatusExpired    SessionStatus = "expired"
)

type TestSession struct {
	ID                   uuid.UUID
	StudentID            uuid.UUID
	TestID               uuid.UUID
	StartedAt            time.Time
	SubmittedAt          *time.Time
	Status               SessionStatus
	TimeRemainingSeconds int
	Answers              []SessionAnswer
	Flags                []SessionQuestionFlag
}

type StatementAnswerInput struct {
	StatementID uuid.UUID
	IsCorrect   bool
}

// SessionAnswer is one row in session_answers.
// For MCQ: SelectedOptionID set, StatementID nil.
// For multi_correct: one row per chosen option, SelectedOptionID set.
// For true_false: one row per statement, StatementID + BooleanAnswer set.
type SessionAnswer struct {
	ID               uuid.UUID
	SessionID        uuid.UUID
	QuestionID       uuid.UUID
	SelectedOptionID *uuid.UUID
	StatementID      *uuid.UUID
	BooleanAnswer    *bool
	AnsweredAt       time.Time
}

type SessionQuestionFlag struct {
	QuestionID uuid.UUID
	IsFlagged  bool
}

type TestResult struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	StudentID    uuid.UUID
	TestID       uuid.UUID
	TotalScore   float64
	CorrectCount int
	WrongCount   int
	BlankCount   int
	CompletedAt  time.Time
}

type TopicBreakdown struct {
	TopicID      uuid.UUID
	TopicName    string
	Total        int
	CorrectCount int
	WrongCount   int
	BlankCount   int
}

type ResultDetail struct {
	TestResult
	Percentage     float64
	TopicBreakdown []TopicBreakdown
}

type AnswerStatus string

const (
	AnswerStatusCorrect AnswerStatus = "correct"
	AnswerStatusWrong   AnswerStatus = "wrong"
	AnswerStatusBlank   AnswerStatus = "blank"
)

type ReviewOption struct {
	ID        uuid.UUID
	Label     string
	Text      string
	IsCorrect bool
}

type ReviewStatement struct {
	ID            uuid.UUID
	Text          string
	IsCorrect     bool  // correct answer
	StudentAnswer *bool // what student picked (nil = not answered)
}

type LeaderboardScope string

const (
	ScopeGlobal LeaderboardScope = "global"
	ScopeTKA    LeaderboardScope = "tka"
	ScopeSMBT   LeaderboardScope = "smbt"
	ScopeWeek   LeaderboardScope = "week"
)

type LeaderboardEntry struct {
	Rank        int64
	StudentID   uuid.UUID
	StudentName string
	TotalScore  float64
	TestCount   int
}

type ReviewItem struct {
	QuestionID   uuid.UUID
	QuestionType QuestionType
	OrderIndex   int
	Text         string
	Explanation  *string
	Difficulty   Difficulty
	TopicID      uuid.UUID
	TopicName    string

	// MCQ
	Options          []ReviewOption
	SelectedOptionID *uuid.UUID
	CorrectOptionID  uuid.UUID

	// multi_correct
	SelectedOptionIDs []uuid.UUID
	CorrectOptionIDs  []uuid.UUID

	// true_false
	Statements []ReviewStatement

	Status AnswerStatus
}
