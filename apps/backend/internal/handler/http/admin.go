package http

import (
	"context"
	"errors"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

func (s *APIServer) GetAdminStats(ctx context.Context, _ api.GetAdminStatsRequestObject) (api.GetAdminStatsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetAdminStats401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.GetAdminStats403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	stats, err := s.adminSvc.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetAdminStats200JSONResponse(api.PlatformStatsResponse{
		TotalStudents:        stats.TotalStudents,
		TotalContributors:    stats.TotalContributors,
		TotalTests:           stats.TotalTests,
		TotalQuestions:       stats.TotalQuestions,
		PendingApprovals:     stats.PendingApprovals,
		StudentsThisWeek:     stats.StudentsThisWeek,
		ContributorsThisWeek: stats.ContributorsThisWeek,
		TotalAttempts:        stats.TotalAttempts,
		AvgScore:             float32(stats.AvgScore),
		TopTestTitle:         stats.TopTestTitle,
		TopTestAttempts:      stats.TopTestAttempts,
		QuestionsTkaSaintek:  stats.QuestionsTKASaintek,
		QuestionsTkaSoshum:   stats.QuestionsTKASoshum,
		QuestionsSmbt:        stats.QuestionsSMBT,
		QuestionsUsed:        stats.QuestionsUsed,
		QuestionsUnused:      stats.QuestionsUnused,
	}), nil
}

func (s *APIServer) ListAdminUsers(ctx context.Context, req api.ListAdminUsersRequestObject) (api.ListAdminUsersResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.ListAdminUsers401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.ListAdminUsers403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	f := repository.UserAdminFilter{
		Page:  derefInt(req.Params.Page, 1),
		Limit: derefInt(req.Params.Limit, 20),
	}
	if req.Params.Search != nil {
		f.Search = *req.Params.Search
	}
	if req.Params.Role != nil {
		r := domain.Role(string(*req.Params.Role))
		f.Role = &r
	}
	if req.Params.Status != nil {
		st := domain.Status(string(*req.Params.Status))
		f.Status = &st
	}

	users, total, err := s.adminSvc.ListUsers(ctx, f)
	if err != nil {
		return nil, err
	}

	resp := api.ListAdminUsers200JSONResponse{
		Data:  make([]api.UserResponse, 0, len(users)),
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	}
	for _, u := range users {
		resp.Data = append(resp.Data, toUserResponse(u))
	}
	return resp, nil
}

func (s *APIServer) UpdateAdminUserStatus(ctx context.Context, req api.UpdateAdminUserStatusRequestObject) (api.UpdateAdminUserStatusResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.UpdateAdminUserStatus401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.UpdateAdminUserStatus403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	status := domain.Status(string(req.Body.Status))
	if err := s.adminSvc.UpdateUserStatus(ctx, req.UserId, status); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.UpdateAdminUserStatus404JSONResponse(errBody("NOT_FOUND", "user not found")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.UpdateAdminUserStatus422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}

	user, err := s.authSvc.Me(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return api.UpdateAdminUserStatus200JSONResponse(toUserResponse(user)), nil
}

func (s *APIServer) AdminCreateContributor(ctx context.Context, req api.AdminCreateContributorRequestObject) (api.AdminCreateContributorResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.AdminCreateContributor401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.AdminCreateContributor403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	user, err := s.adminSvc.CreateContributor(ctx, req.Body.Name, string(req.Body.Email), req.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrConflict):
			return api.AdminCreateContributor409JSONResponse(errBody("CONFLICT", "email already registered")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.AdminCreateContributor400JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.AdminCreateContributor201JSONResponse(toUserResponse(user)), nil
}

func (s *APIServer) ListPendingContributors(ctx context.Context, req api.ListPendingContributorsRequestObject) (api.ListPendingContributorsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.ListPendingContributors401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.ListPendingContributors403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	roleContrib := domain.RoleContributor
	statusPending := domain.StatusPending
	f := repository.UserAdminFilter{
		Role:   &roleContrib,
		Status: &statusPending,
		Page:   derefInt(req.Params.Page, 1),
		Limit:  derefInt(req.Params.Limit, 20),
	}

	users, total, err := s.adminSvc.ListUsers(ctx, f)
	if err != nil {
		return nil, err
	}

	resp := api.ListPendingContributors200JSONResponse{
		Data:  make([]api.UserResponse, 0, len(users)),
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	}
	for _, u := range users {
		resp.Data = append(resp.Data, toUserResponse(u))
	}
	return resp, nil
}

func (s *APIServer) ApproveContributor(ctx context.Context, req api.ApproveContributorRequestObject) (api.ApproveContributorResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.ApproveContributor401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.ApproveContributor403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	if err := s.adminSvc.ApproveContributor(ctx, req.UserId); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.ApproveContributor404JSONResponse(errBody("NOT_FOUND", "user not found")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.ApproveContributor422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.ApproveContributor200JSONResponse(api.MessageResponse{Message: "contributor approved"}), nil
}

func (s *APIServer) RejectContributor(ctx context.Context, req api.RejectContributorRequestObject) (api.RejectContributorResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.RejectContributor401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.RejectContributor403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	if err := s.adminSvc.RejectContributor(ctx, req.UserId); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.RejectContributor404JSONResponse(errBody("NOT_FOUND", "user not found")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.RejectContributor422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.RejectContributor200JSONResponse(api.MessageResponse{Message: "contributor rejected"}), nil
}

func (s *APIServer) ListAdminTests(ctx context.Context, req api.ListAdminTestsRequestObject) (api.ListAdminTestsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.ListAdminTests401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.ListAdminTests403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	page := derefInt(req.Params.Page, 1)
	limit := derefInt(req.Params.Limit, 20)

	var eduLevel *domain.EducationLevel
	if req.Params.EducationLevel != nil {
		el := domain.EducationLevel(string(*req.Params.EducationLevel))
		eduLevel = &el
	}

	tests, total, err := s.adminSvc.ListTestsWithAttempts(ctx, page, limit, eduLevel)
	if err != nil {
		return nil, err
	}

	resp := api.ListAdminTests200JSONResponse{
		Data:  make([]api.AdminTestResponse, 0, len(tests)),
		Total: total,
		Page:  page,
		Limit: limit,
	}
	for _, t := range tests {
		resp.Data = append(resp.Data, toAdminTestResponse(t))
	}
	return resp, nil
}

func toAdminTestResponse(t *domain.TestWithAttempts) api.AdminTestResponse {
	resp := api.AdminTestResponse{
		Id:              t.ID,
		ContributorId:   t.ContributorID,
		Title:           t.Title,
		Category:        api.AdminTestResponseCategory(t.Category),
		Difficulty:      api.AdminTestResponseDifficulty(t.Difficulty),
		Status:          api.AdminTestResponseStatus(t.Status),
		DurationMinutes: t.DurationMinutes,
		CreatedAt:       t.CreatedAt,
		AttemptCount:    t.AttemptCount,
	}
	if t.EducationLevel != nil {
		el := api.AdminTestResponseEducationLevel(string(*t.EducationLevel))
		resp.EducationLevel = &el
	}
	return resp
}

func derefInt(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}
