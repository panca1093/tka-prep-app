package http

import (
	"context"
	"errors"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
)

func (s *APIServer) GetLeaderboard(ctx context.Context, req api.GetLeaderboardRequestObject) (api.GetLeaderboardResponseObject, error) {
	_, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetLeaderboard401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	scope := resolveScope((*string)(req.Params.Scope))
	entries, err := s.leaderboardRepo.List(ctx, scope)
	if err != nil {
		return nil, err
	}

	resp := api.GetLeaderboard200JSONResponse{Data: make([]api.LeaderboardEntryResponse, 0, len(entries))}
	for _, e := range entries {
		resp.Data = append(resp.Data, toLeaderboardEntryResponse(e))
	}
	return resp, nil
}

func (s *APIServer) GetLeaderboardMe(ctx context.Context, req api.GetLeaderboardMeRequestObject) (api.GetLeaderboardMeResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetLeaderboardMe401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleStudent {
		return api.GetLeaderboardMe403JSONResponse(errBody("FORBIDDEN", "student only")), nil
	}

	scope := resolveScope((*string)(req.Params.Scope))
	entry, err := s.leaderboardRepo.GetMyRank(ctx, claims.UserID, scope)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return api.GetLeaderboardMe404JSONResponse(errBody("NOT_FOUND", "no results found for this scope")), nil
		}
		return nil, err
	}

	return api.GetLeaderboardMe200JSONResponse(toLeaderboardEntryResponse(*entry)), nil
}

func resolveScope(s *string) domain.LeaderboardScope {
	if s == nil {
		return domain.ScopeGlobal
	}
	switch domain.LeaderboardScope(*s) {
	case domain.ScopeTKA, domain.ScopeSMBT, domain.ScopeWeek:
		return domain.LeaderboardScope(*s)
	default:
		return domain.ScopeGlobal
	}
}

func toLeaderboardEntryResponse(e domain.LeaderboardEntry) api.LeaderboardEntryResponse {
	return api.LeaderboardEntryResponse{
		Rank:        e.Rank,
		StudentId:   e.StudentID,
		StudentName: e.StudentName,
		TotalScore:  e.TotalScore,
		TestCount:   e.TestCount,
	}
}
