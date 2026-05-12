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
	Text          string
	Explanation   *string
	Difficulty    Difficulty
	Options       []QuestionOption
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
