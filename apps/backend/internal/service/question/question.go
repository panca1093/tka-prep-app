package question

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
	questions repository.QuestionRepository
}

func NewService(questions repository.QuestionRepository) *Service {
	return &Service{questions: questions}
}

type OptionInput struct {
	Label     string
	Text      string
	IsCorrect bool
}

type CreateInput struct {
	ContributorID uuid.UUID
	TopicID       uuid.UUID
	Text          string
	Explanation   *string
	Difficulty    domain.Difficulty
	Options       []OptionInput
}

type UpdateInput struct {
	TopicID     *uuid.UUID
	Text        *string
	Explanation *string
	Difficulty  *domain.Difficulty
	Options     []OptionInput
}

type ListFilter struct {
	Search     string
	TopicID    *uuid.UUID
	Difficulty *domain.Difficulty
	Page       int
	Limit      int
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*domain.Question, int, error) {
	return s.questions.List(ctx, repository.QuestionFilter{
		Search:     f.Search,
		TopicID:    f.TopicID,
		Difficulty: f.Difficulty,
		Page:       f.Page,
		Limit:      f.Limit,
	})
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Question, error) {
	if err := validateOptions(in.Options); err != nil {
		return nil, err
	}

	in.Text = strings.TrimSpace(in.Text)
	if in.Text == "" {
		return nil, fmt.Errorf("%w: question text is required", apierr.ErrValidation)
	}

	now := time.Now().UTC()
	q := &domain.Question{
		ID:            uuid.New(),
		ContributorID: in.ContributorID,
		TopicID:       in.TopicID,
		Text:          in.Text,
		Explanation:   in.Explanation,
		Difficulty:    in.Difficulty,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range in.Options {
		q.Options = append(q.Options, domain.QuestionOption{
			ID:        uuid.New(),
			Label:     o.Label,
			Text:      o.Text,
			IsCorrect: o.IsCorrect,
		})
	}

	if err := s.questions.Create(ctx, q); err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}
	return q, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	return s.questions.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role, in UpdateInput) (*domain.Question, error) {
	q, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && q.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}

	if in.Text != nil {
		text := strings.TrimSpace(*in.Text)
		if text == "" {
			return nil, fmt.Errorf("%w: question text cannot be empty", apierr.ErrValidation)
		}
		q.Text = text
	}
	if in.TopicID != nil {
		q.TopicID = *in.TopicID
	}
	if in.Explanation != nil {
		q.Explanation = in.Explanation
	}
	if in.Difficulty != nil {
		q.Difficulty = *in.Difficulty
	}
	if len(in.Options) > 0 {
		if err := validateOptions(in.Options); err != nil {
			return nil, err
		}
		q.Options = make([]domain.QuestionOption, len(in.Options))
		for i, o := range in.Options {
			q.Options[i] = domain.QuestionOption{
				ID:        uuid.New(),
				Label:     o.Label,
				Text:      o.Text,
				IsCorrect: o.IsCorrect,
			}
		}
	}
	q.UpdatedAt = time.Now().UTC()

	if err := s.questions.Update(ctx, q); err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}
	return q, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role) error {
	q, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if callerRole != domain.RoleAdmin && q.ContributorID != callerID {
		return apierr.ErrForbidden
	}

	used, err := s.questions.IsUsedInPublishedTest(ctx, id)
	if err != nil {
		return fmt.Errorf("check question usage: %w", err)
	}
	if used {
		return apierr.ErrConflict
	}

	return s.questions.Delete(ctx, id)
}

func validateOptions(opts []OptionInput) error {
	if len(opts) != 5 {
		return fmt.Errorf("%w: exactly 5 options are required", apierr.ErrValidation)
	}

	labels := map[string]bool{}
	correctCount := 0
	for _, o := range opts {
		label := strings.ToUpper(strings.TrimSpace(o.Label))
		if label == "" || !strings.ContainsAny(label, "ABCDE") || len(label) != 1 {
			return fmt.Errorf("%w: option label must be A, B, C, D, or E", apierr.ErrValidation)
		}
		if labels[label] {
			return fmt.Errorf("%w: duplicate option label %s", apierr.ErrValidation, label)
		}
		labels[label] = true
		if strings.TrimSpace(o.Text) == "" {
			return fmt.Errorf("%w: option text cannot be empty", apierr.ErrValidation)
		}
		if o.IsCorrect {
			correctCount++
		}
	}

	if correctCount != 1 {
		return fmt.Errorf("%w: exactly one option must be marked correct", apierr.ErrValidation)
	}

	for _, required := range []string{"A", "B", "C", "D", "E"} {
		if !labels[required] {
			return fmt.Errorf("%w: missing option %s", apierr.ErrValidation, required)
		}
	}

	return nil
}

// EnsureContributorActive is called before create to ensure the contributor is active (not pending).
func EnsureContributorActive(status domain.Status) error {
	if status != domain.StatusActive {
		return errors.New("contributor account is not active")
	}
	return nil
}
