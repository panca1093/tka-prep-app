package topic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type Service struct {
	topics repository.TopicRepository
}

func NewService(topics repository.TopicRepository) *Service {
	return &Service{topics: topics}
}

func (s *Service) List(ctx context.Context) ([]*domain.Topic, error) {
	return s.topics.List(ctx)
}

type CreateInput struct {
	Name        string
	Description *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Topic, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, fmt.Errorf("%w: name is required", apierr.ErrValidation)
	}

	t := &domain.Topic{
		ID:          uuid.New(),
		Name:        in.Name,
		Description: in.Description,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.topics.Create(ctx, t); err != nil {
		if errors.Is(err, apierr.ErrConflict) {
			return nil, apierr.ErrConflict
		}
		return nil, fmt.Errorf("create topic: %w", err)
	}
	return t, nil
}

type UpdateInput struct {
	Name        *string
	Description *string
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (*domain.Topic, error) {
	t, err := s.topics.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", apierr.ErrValidation)
		}
		t.Name = name
	}
	if in.Description != nil {
		t.Description = in.Description
	}

	if err := s.topics.Update(ctx, t); err != nil {
		if errors.Is(err, apierr.ErrConflict) {
			return nil, apierr.ErrConflict
		}
		return nil, fmt.Errorf("update topic: %w", err)
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	used, err := s.topics.IsUsedByQuestions(ctx, id)
	if err != nil {
		return fmt.Errorf("check topic usage: %w", err)
	}
	if used {
		return apierr.ErrConflict
	}
	return s.topics.Delete(ctx, id)
}
