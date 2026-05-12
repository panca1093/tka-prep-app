package domain

import (
	"time"

	"github.com/google/uuid"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type QuestionType string

const (
	QuestionTypeMCQ          QuestionType = "mcq"
	QuestionTypeTrueFalse    QuestionType = "true_false"
	QuestionTypeMultiCorrect QuestionType = "multi_correct"
)

type Topic struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
}

type Question struct {
	ID            uuid.UUID
	ContributorID uuid.UUID
	TopicID       uuid.UUID
	Type          QuestionType
	Text          string
	Explanation   *string
	Difficulty    Difficulty
	Options       []QuestionOption   // MCQ and multi_correct
	Statements    []QuestionStatement // true_false only
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type QuestionOption struct {
	ID         uuid.UUID
	QuestionID uuid.UUID
	Label      string
	Text       string
	IsCorrect  bool
}

type QuestionStatement struct {
	ID         uuid.UUID
	QuestionID uuid.UUID
	Text       string
	IsCorrect  bool
	Position   int
}
