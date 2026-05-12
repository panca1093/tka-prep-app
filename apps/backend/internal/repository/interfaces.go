package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
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
	Search     string
	TopicID    *uuid.UUID
	Difficulty *domain.Difficulty
	Page       int
	Limit      int
}

type QuestionRepository interface {
	List(ctx context.Context, f QuestionFilter) ([]*domain.Question, int, error)
	Create(ctx context.Context, q *domain.Question) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	Update(ctx context.Context, q *domain.Question) error
	Delete(ctx context.Context, id uuid.UUID) error
	// IsUsedInPublishedTest reports whether the question is referenced by any published test.
	IsUsedInPublishedTest(ctx context.Context, id uuid.UUID) (bool, error)
}
