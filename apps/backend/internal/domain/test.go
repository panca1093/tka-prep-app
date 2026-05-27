package domain

import (
	"time"

	"github.com/google/uuid"
)

type TestStatus string

const (
	TestStatusDraft     TestStatus = "draft"
	TestStatusPublished TestStatus = "published"
)

type Test struct {
	ID              uuid.UUID
	ContributorID   uuid.UUID
	Title           string
	Description     *string
	CategoryID      uuid.UUID
	CategoryName    string
	DurationMinutes int
	Difficulty      Difficulty
	Status          TestStatus
	EducationLevel  *EducationLevel
	CreatedAt       time.Time
	PublishedAt     *time.Time
	Questions       []TestQuestion
	ScoringConfig   *ScoringConfig
	// Populated only for student callers — computed from test_sessions + test_results.
	StudentStatus string     // "not_started" | "in_progress" | "completed"
	ResultID      *uuid.UUID // set when StudentStatus == "completed"
	// Populated for contributor callers — computed from test_sessions + test_results.
	AttemptCount int      // number of submitted attempts
	AvgScore     *float64 // null when no submissions
}

type TestQuestion struct {
	ID         uuid.UUID
	TestID     uuid.UUID
	QuestionID uuid.UUID
	OrderIndex int
}

type ScoringConfig struct {
	ID            uuid.UUID
	TestID        uuid.UUID
	CorrectPoints float64
	WrongPoints   float64
	BlankPoints   float64
}
