package http

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	testsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/test"
)

func (s *APIServer) GetTests(ctx context.Context, req api.GetTestsRequestObject) (api.GetTestsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetTests401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	p := req.Params
	f := testsvc.ListFilter{Page: 1, Limit: 20}
	if p.Page != nil {
		f.Page = *p.Page
	}
	if p.Limit != nil {
		f.Limit = *p.Limit
	}
	if p.CategoryId != nil {
		id := uuid.UUID(*p.CategoryId)
		f.CategoryID = &id
	}
	if p.Difficulty != nil {
		d := domain.Difficulty(*p.Difficulty)
		f.Difficulty = &d
	}
	if p.Status != nil {
		st := domain.TestStatus(*p.Status)
		f.Status = &st
	}
	if p.ContributorId != nil {
		f.ContributorID = p.ContributorId
	}

	// Students only see published tests; contributors see own draft+published; admin sees all.
	if claims.Role == domain.RoleStudent {
		published := domain.TestStatusPublished
		f.Status = &published
		// Filter tests by student's education level.
		user, err := s.authSvc.Me(ctx, claims.UserID)
		if err == nil && user.EducationLevel != nil {
			f.EducationLevel = user.EducationLevel
		}
		f.StudentID = &claims.UserID
	} else if claims.Role == domain.RoleContributor && f.ContributorID == nil {
		f.ContributorID = &claims.UserID
	}

	tests, total, err := s.testSvc.List(ctx, f)
	if err != nil {
		return nil, err
	}

	resp := api.GetTests200JSONResponse{
		Data:  make([]api.TestDetailResponse, 0, len(tests)),
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	}
	for _, t := range tests {
		resp.Data = append(resp.Data, toTestDetailResponse(t))
	}
	return resp, nil
}

func (s *APIServer) PostTests(ctx context.Context, req api.PostTestsRequestObject) (api.PostTestsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostTests401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PostTests403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	in := testsvc.CreateInput{
		ContributorID:   claims.UserID,
		Title:           req.Body.Title,
		Description:     req.Body.Description,
		CategoryID:      req.Body.CategoryId,
		DurationMinutes: req.Body.DurationMinutes,
		Difficulty:      domain.Difficulty(req.Body.Difficulty),
	}
	if req.Body.EducationLevel != nil {
		el := domain.EducationLevel(string(*req.Body.EducationLevel))
		in.EducationLevel = &el
	}
	if req.Body.ScoringConfig != nil {
		in.ScoringConfig = &testsvc.ScoringConfigInput{
			CorrectPoints: req.Body.ScoringConfig.CorrectPoints,
			WrongPoints:   req.Body.ScoringConfig.WrongPoints,
			BlankPoints:   req.Body.ScoringConfig.BlankPoints,
		}
	}

	t, err := s.testSvc.Create(ctx, in)
	if err != nil {
		if errors.Is(err, apierr.ErrValidation) {
			return api.PostTests422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.PostTests201JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) GetTestsTestId(ctx context.Context, req api.GetTestsTestIdRequestObject) (api.GetTestsTestIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetTestsTestId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	t, err := s.testSvc.Get(ctx, req.TestId)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return api.GetTestsTestId404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		}
		return nil, err
	}

	// Students can only see published tests.
	if claims.Role == domain.RoleStudent && t.Status != domain.TestStatusPublished {
		return api.GetTestsTestId404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
	}
	// Students cannot access tests outside their education level.
	if claims.Role == domain.RoleStudent && t.EducationLevel != nil {
		user, err := s.authSvc.Me(ctx, claims.UserID)
		if err == nil && user.EducationLevel != nil && *t.EducationLevel != *user.EducationLevel {
			return api.GetTestsTestId403JSONResponse(errBody("FORBIDDEN", "this test is not available for your education level")), nil
		}
	}
	// Contributors can only see their own drafts.
	if claims.Role == domain.RoleContributor && t.Status == domain.TestStatusDraft && t.ContributorID != claims.UserID {
		return api.GetTestsTestId403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
	}

	return api.GetTestsTestId200JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) PatchTestsTestId(ctx context.Context, req api.PatchTestsTestIdRequestObject) (api.PatchTestsTestIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PatchTestsTestId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PatchTestsTestId403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	in := testsvc.UpdateInput{
		Title:       req.Body.Title,
		Description: req.Body.Description,
	}
	if req.Body.CategoryId != nil {
		in.CategoryID = req.Body.CategoryId
	}
	if req.Body.DurationMinutes != nil {
		in.DurationMinutes = req.Body.DurationMinutes
	}
	if req.Body.Difficulty != nil {
		d := domain.Difficulty(*req.Body.Difficulty)
		in.Difficulty = &d
	}
	if req.Body.EducationLevel != nil {
		el := domain.EducationLevel(string(*req.Body.EducationLevel))
		in.EducationLevel = &el
	}

	t, err := s.testSvc.Update(ctx, req.TestId, claims.UserID, claims.Role, in)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PatchTestsTestId404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PatchTestsTestId403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PatchTestsTestId409JSONResponse(errBody("CONFLICT", err.Error())), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PatchTestsTestId422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.PatchTestsTestId200JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) DeleteTestsTestId(ctx context.Context, req api.DeleteTestsTestIdRequestObject) (api.DeleteTestsTestIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.DeleteTestsTestId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.DeleteTestsTestId403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	if err := s.testSvc.Delete(ctx, req.TestId, claims.UserID, claims.Role); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.DeleteTestsTestId404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.DeleteTestsTestId403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.DeleteTestsTestId409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.DeleteTestsTestId200JSONResponse{Message: "deleted"}, nil
}

func (s *APIServer) PostTestsTestIdPublish(ctx context.Context, req api.PostTestsTestIdPublishRequestObject) (api.PostTestsTestIdPublishResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostTestsTestIdPublish401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PostTestsTestIdPublish403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	t, err := s.testSvc.Publish(ctx, req.TestId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PostTestsTestIdPublish404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PostTestsTestIdPublish403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PostTestsTestIdPublish409JSONResponse(errBody("CONFLICT", err.Error())), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PostTestsTestIdPublish422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.PostTestsTestIdPublish200JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) PostTestsTestIdUnpublish(ctx context.Context, req api.PostTestsTestIdUnpublishRequestObject) (api.PostTestsTestIdUnpublishResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostTestsTestIdUnpublish401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	// Only admin can unpublish.
	if claims.Role != domain.RoleAdmin && claims.Role != domain.RoleContributor {
		return api.PostTestsTestIdUnpublish403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	t, err := s.testSvc.Unpublish(ctx, req.TestId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PostTestsTestIdUnpublish404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PostTestsTestIdUnpublish403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PostTestsTestIdUnpublish409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.PostTestsTestIdUnpublish200JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) PutTestsTestIdQuestions(ctx context.Context, req api.PutTestsTestIdQuestionsRequestObject) (api.PutTestsTestIdQuestionsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PutTestsTestIdQuestions401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PutTestsTestIdQuestions403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	t, err := s.testSvc.SetQuestions(ctx, req.TestId, claims.UserID, claims.Role, req.Body.QuestionIds)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PutTestsTestIdQuestions404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PutTestsTestIdQuestions403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PutTestsTestIdQuestions409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.PutTestsTestIdQuestions200JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) PatchTestsTestIdScoring(ctx context.Context, req api.PatchTestsTestIdScoringRequestObject) (api.PatchTestsTestIdScoringResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PatchTestsTestIdScoring401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PatchTestsTestIdScoring403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	t, err := s.testSvc.UpdateScoringConfig(ctx, req.TestId, claims.UserID, claims.Role, testsvc.ScoringConfigInput{
		CorrectPoints: req.Body.CorrectPoints,
		WrongPoints:   req.Body.WrongPoints,
		BlankPoints:   req.Body.BlankPoints,
	})
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PatchTestsTestIdScoring404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PatchTestsTestIdScoring403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PatchTestsTestIdScoring409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.PatchTestsTestIdScoring200JSONResponse(toTestDetailResponse(t)), nil
}

func (s *APIServer) GetTestsTestIdResults(ctx context.Context, req api.GetTestsTestIdResultsRequestObject) (api.GetTestsTestIdResultsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetTestsTestIdResults401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.GetTestsTestIdResults403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	page, limit := 1, 20
	if req.Params.Page != nil {
		page = *req.Params.Page
	}
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	results, total, analytics, err := s.resultSvc.ListByTestID(ctx, req.TestId, claims.UserID, claims.Role, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.GetTestsTestIdResults404JSONResponse(errBody("NOT_FOUND", "test not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.GetTestsTestIdResults403JSONResponse(errBody("FORBIDDEN", "not your test")), nil
		}
		return nil, err
	}

	data := make([]api.ContributorResultEntry, 0, len(results))
	for _, r := range results {
		data = append(data, api.ContributorResultEntry{
			Id:           r.ID,
			SessionId:    r.SessionID,
			StudentId:    r.StudentID,
			StudentName:  r.StudentName,
			StudentEmail: openapi_types.Email(r.StudentEmail),
			TestId:       r.TestID,
			TestTitle:    r.TestTitle,
			TotalScore:   r.TotalScore,
			CorrectCount: r.CorrectCount,
			WrongCount:   r.WrongCount,
			BlankCount:   r.BlankCount,
			CompletedAt:  r.CompletedAt,
		})
	}

	resp := api.GetTestsTestIdResults200JSONResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	if analytics != nil {
		perTopic := make([]struct {
			AvgCorrectCount *float64            `json:"avg_correct_count,omitempty"`
			QuestionCount   *int                `json:"question_count,omitempty"`
			TopicId         *openapi_types.UUID `json:"topic_id,omitempty"`
			TopicName       *string             `json:"topic_name,omitempty"`
		}, 0, len(analytics.PerTopic))
		for _, pt := range analytics.PerTopic {
			ac := pt.AvgCorrectCount
			qc := pt.QuestionCount
			tid := openapi_types.UUID(pt.TopicID)
			tn := pt.TopicName
			perTopic = append(perTopic, struct {
				AvgCorrectCount *float64            `json:"avg_correct_count,omitempty"`
				QuestionCount   *int                `json:"question_count,omitempty"`
				TopicId         *openapi_types.UUID `json:"topic_id,omitempty"`
				TopicName       *string             `json:"topic_name,omitempty"`
			}{
				AvgCorrectCount: &ac,
				QuestionCount:   &qc,
				TopicId:         &tid,
				TopicName:       &tn,
			})
		}
		ta := int(analytics.TotalAttempts)
		var analyticsResp api.TestAnalyticsResponse
		analyticsResp.TotalAttempts = &ta
		analyticsResp.AvgScore = analytics.AvgScore
		analyticsResp.MaxScore = analytics.MaxScore
		analyticsResp.MinScore = analytics.MinScore
		analyticsResp.PerTopic = &perTopic
		resp.Analytics = &analyticsResp
	}

	return resp, nil
}

func toTestDetailResponse(t *domain.Test) api.TestDetailResponse {
	qs := make([]api.TestQuestionResponse, len(t.Questions))
	for i, q := range t.Questions {
		qs[i] = api.TestQuestionResponse{
			Id:         q.ID,
			TestId:     q.TestID,
			QuestionId: q.QuestionID,
			OrderIndex: q.OrderIndex,
		}
	}

	resp := api.TestDetailResponse{
		Id:              t.ID,
		ContributorId:   t.ContributorID,
		Title:           t.Title,
		Description:     t.Description,
		CategoryId:      t.CategoryID,
			CategoryName:    t.CategoryName,
		DurationMinutes: t.DurationMinutes,
		Difficulty:      api.TestDetailResponseDifficulty(t.Difficulty),
		Status:          api.TestDetailResponseStatus(t.Status),
		CreatedAt:       t.CreatedAt,
		PublishedAt:     t.PublishedAt,
		Questions:       qs,
	}
	if t.EducationLevel != nil {
		el := api.TestDetailResponseEducationLevel(string(*t.EducationLevel))
		resp.EducationLevel = &el
	}
	if t.AttemptCount > 0 || t.AvgScore != nil {
		resp.AttemptCount = &t.AttemptCount
		resp.AvgScore = t.AvgScore
	}
	if t.StudentStatus != "" {
		st := api.TestDetailResponseStudentStatus(t.StudentStatus)
		resp.StudentStatus = &st
		if t.ResultID != nil {
			resp.ResultId = t.ResultID
		}
	}

	if t.ScoringConfig != nil {
		sc := api.ScoringConfigResponse{
			Id:            t.ScoringConfig.ID,
			TestId:        t.ScoringConfig.TestID,
			CorrectPoints: t.ScoringConfig.CorrectPoints,
			WrongPoints:   t.ScoringConfig.WrongPoints,
			BlankPoints:   t.ScoringConfig.BlankPoints,
		}
		resp.ScoringConfig = &sc
	}

	return resp
}
