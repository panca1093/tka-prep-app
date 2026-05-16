package domain

import (
	"time"

	"github.com/google/uuid"
)

type TestCategory string

const (
	CategoryTKASaintek TestCategory = "tka_saintek"
	CategoryTKASoshum  TestCategory = "tka_soshum"
	CategorySMBT       TestCategory = "smbt"
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
	Category        TestCategory
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
