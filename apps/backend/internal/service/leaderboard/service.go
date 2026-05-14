package leaderboard

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type Service struct {
	repo repository.LeaderboardRepository
}

func NewService(repo repository.LeaderboardRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, scope domain.LeaderboardScope) ([]domain.LeaderboardEntry, error) {
	return s.repo.List(ctx, scope)
}

func (s *Service) GetMyRank(ctx context.Context, studentID uuid.UUID, scope domain.LeaderboardScope) (*domain.LeaderboardEntry, error) {
	return s.repo.GetMyRank(ctx, studentID, scope)
}
