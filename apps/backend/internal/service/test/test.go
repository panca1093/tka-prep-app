package test

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
	tests repository.TestRepository
}

func NewService(tests repository.TestRepository) *Service {
	return &Service{tests: tests}
}

type CreateInput struct {
	ContributorID   uuid.UUID
	Title           string
	Description     *string
	CategoryID      uuid.UUID
	DurationMinutes int
	Difficulty      domain.Difficulty
	EducationLevel  *domain.EducationLevel
	ScoringConfig   *ScoringConfigInput
}

type ScoringConfigInput struct {
	CorrectPoints float64
	WrongPoints   float64
	BlankPoints   float64
}

type UpdateInput struct {
	Title           *string
	Description     *string
	CategoryID      *uuid.UUID
	DurationMinutes *int
	Difficulty      *domain.Difficulty
	EducationLevel  *domain.EducationLevel
}

type SetQuestionsInput struct {
	QuestionIDs []uuid.UUID
}

type ListFilter struct {
	ContributorID  *uuid.UUID
	CategoryID     *uuid.UUID
	Difficulty     *domain.Difficulty
	Status         *domain.TestStatus
	EducationLevel *domain.EducationLevel
	StudentID      *uuid.UUID
	Page           int
	Limit          int
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Test, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return nil, fmt.Errorf("%w: title is required", apierr.ErrValidation)
	}
	if in.DurationMinutes <= 0 {
		return nil, fmt.Errorf("%w: duration_minutes must be positive", apierr.ErrValidation)
	}

	now := time.Now().UTC()
	t := &domain.Test{
		ID:              uuid.New(),
		ContributorID:   in.ContributorID,
		Title:           in.Title,
		Description:     in.Description,
		CategoryID:      in.CategoryID,
		DurationMinutes: in.DurationMinutes,
		Difficulty:      in.Difficulty,
		EducationLevel:  in.EducationLevel,
		Status:          domain.TestStatusDraft,
		CreatedAt:       now,
	}

	if in.ScoringConfig != nil {
		t.ScoringConfig = &domain.ScoringConfig{
			ID:            uuid.New(),
			TestID:        t.ID,
			CorrectPoints: in.ScoringConfig.CorrectPoints,
			WrongPoints:   in.ScoringConfig.WrongPoints,
			BlankPoints:   in.ScoringConfig.BlankPoints,
		}
	} else {
		t.ScoringConfig = &domain.ScoringConfig{
			ID:            uuid.New(),
			TestID:        t.ID,
			CorrectPoints: 4,
			WrongPoints:   0,
			BlankPoints:   0,
		}
	}

	if err := s.tests.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create test: %w", err)
	}
	return t, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Test, error) {
	return s.tests.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*domain.Test, int, error) {
	return s.tests.List(ctx, repository.TestFilter{
		ContributorID:  f.ContributorID,
		CategoryID:     f.CategoryID,
		Difficulty:     f.Difficulty,
		Status:         f.Status,
		EducationLevel: f.EducationLevel,
		StudentID:      f.StudentID,
		Page:           f.Page,
		Limit:          f.Limit,
	})
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role, in UpdateInput) (*domain.Test, error) {
	t, err := s.tests.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && t.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}
	if t.Status == domain.TestStatusPublished {
		return nil, fmt.Errorf("%w: cannot edit a published test", apierr.ErrConflict)
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", apierr.ErrValidation)
		}
		t.Title = title
	}
	if in.Description != nil {
		t.Description = in.Description
	}
	if in.CategoryID != nil {
		t.CategoryID = *in.CategoryID
	}
	if in.DurationMinutes != nil {
		if *in.DurationMinutes <= 0 {
			return nil, fmt.Errorf("%w: duration_minutes must be positive", apierr.ErrValidation)
		}
		t.DurationMinutes = *in.DurationMinutes
	}
	if in.Difficulty != nil {
		t.Difficulty = *in.Difficulty
	}
	if in.EducationLevel != nil {
		t.EducationLevel = in.EducationLevel
	}

	if err := s.tests.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update test: %w", err)
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role) error {
	t, err := s.tests.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if callerRole != domain.RoleAdmin && t.ContributorID != callerID {
		return apierr.ErrForbidden
	}
	if t.Status == domain.TestStatusPublished {
		return fmt.Errorf("%w: cannot delete a published test", apierr.ErrConflict)
	}

	return s.tests.Delete(ctx, id)
}

func (s *Service) Publish(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role) (*domain.Test, error) {
	t, err := s.tests.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && t.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}
	if t.Status == domain.TestStatusPublished {
		return nil, fmt.Errorf("%w: test is already published", apierr.ErrConflict)
	}

	count, err := s.tests.QuestionCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("check question count: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("%w: test must have at least one question before publishing", apierr.ErrValidation)
	}

	now := time.Now().UTC()
	if err := s.tests.SetStatus(ctx, id, domain.TestStatusPublished, &now); err != nil {
		return nil, fmt.Errorf("publish test: %w", err)
	}
	t.Status = domain.TestStatusPublished
	t.PublishedAt = &now
	return t, nil
}

func (s *Service) Unpublish(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role) (*domain.Test, error) {
	t, err := s.tests.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if callerRole == domain.RoleContributor && t.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}

	if t.Status == domain.TestStatusDraft {
		return nil, fmt.Errorf("%w: test is already a draft", apierr.ErrConflict)
	}

	if err := s.tests.SetStatus(ctx, id, domain.TestStatusDraft, nil); err != nil {
		return nil, fmt.Errorf("unpublish test: %w", err)
	}
	t.Status = domain.TestStatusDraft
	t.PublishedAt = nil
	return t, nil
}

func (s *Service) SetQuestions(ctx context.Context, testID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, questionIDs []uuid.UUID) (*domain.Test, error) {
	t, err := s.tests.FindByID(ctx, testID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && t.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}
	if t.Status == domain.TestStatusPublished {
		return nil, fmt.Errorf("%w: cannot modify questions of a published test", apierr.ErrConflict)
	}

	questions := make([]domain.TestQuestion, len(questionIDs))
	for i, qID := range questionIDs {
		questions[i] = domain.TestQuestion{
			ID:         uuid.New(),
			TestID:     testID,
			QuestionID: qID,
			OrderIndex: i + 1,
		}
	}

	if err := s.tests.SetQuestions(ctx, testID, questions); err != nil {
		return nil, fmt.Errorf("set questions: %w", err)
	}
	t.Questions = questions
	return t, nil
}

func (s *Service) UpdateScoringConfig(ctx context.Context, testID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, in ScoringConfigInput) (*domain.Test, error) {
	t, err := s.tests.FindByID(ctx, testID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && t.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}
	if t.Status == domain.TestStatusPublished {
		return nil, fmt.Errorf("%w: cannot modify scoring of a published test", apierr.ErrConflict)
	}

	sc := &domain.ScoringConfig{
		TestID:        testID,
		CorrectPoints: in.CorrectPoints,
		WrongPoints:   in.WrongPoints,
		BlankPoints:   in.BlankPoints,
	}
	if t.ScoringConfig != nil {
		sc.ID = t.ScoringConfig.ID
	}

	if err := s.tests.UpdateScoringConfig(ctx, sc); err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return nil, fmt.Errorf("%w: scoring config not found for this test", apierr.ErrNotFound)
		}
		return nil, fmt.Errorf("update scoring config: %w", err)
	}
	t.ScoringConfig = sc
	return t, nil
}
