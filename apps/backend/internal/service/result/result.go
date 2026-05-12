package result

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type ReviewFilter string

const (
	ReviewFilterAll     ReviewFilter = "all"
	ReviewFilterCorrect ReviewFilter = "correct"
	ReviewFilterWrong   ReviewFilter = "wrong"
	ReviewFilterBlank   ReviewFilter = "blank"
)

type Service struct {
	results  repository.ResultRepository
	sessions repository.SessionRepository
}

func NewService(results repository.ResultRepository, sessions repository.SessionRepository) *Service {
	return &Service{results: results, sessions: sessions}
}

// List returns results scoped by caller role: students see own, admin sees all.
func (s *Service) List(ctx context.Context, callerID uuid.UUID, callerRole domain.Role) ([]*domain.TestResult, error) {
	if callerRole == domain.RoleAdmin {
		return s.results.ListAll(ctx)
	}
	return s.results.ListByStudent(ctx, callerID)
}

// Get returns the detailed result. Students can only access their own; admin can access any.
func (s *Service) Get(ctx context.Context, resultID, callerID uuid.UUID, callerRole domain.Role) (*domain.ResultDetail, error) {
	detail, err := s.results.FindDetailByID(ctx, resultID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && detail.StudentID != callerID {
		return nil, apierr.ErrForbidden
	}
	return detail, nil
}

// GetReview returns per-question review, optionally filtered by answer status.
func (s *Service) GetReview(ctx context.Context, resultID, callerID uuid.UUID, callerRole domain.Role, filter ReviewFilter) ([]domain.ReviewItem, error) {
	result, err := s.results.FindByID(ctx, resultID)
	if err != nil {
		return nil, err
	}

	if callerRole != domain.RoleAdmin && result.StudentID != callerID {
		return nil, apierr.ErrForbidden
	}

	items, err := s.results.GetReview(ctx, result.SessionID, result.TestID)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}

	if filter == ReviewFilterAll || filter == "" {
		return items, nil
	}

	var target domain.AnswerStatus
	switch filter {
	case ReviewFilterCorrect:
		target = domain.AnswerStatusCorrect
	case ReviewFilterWrong:
		target = domain.AnswerStatusWrong
	case ReviewFilterBlank:
		target = domain.AnswerStatusBlank
	default:
		return nil, fmt.Errorf("%w: invalid filter %q; use all, correct, wrong, or blank", apierr.ErrValidation, filter)
	}

	filtered := items[:0]
	for _, item := range items {
		if item.Status == target {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}
