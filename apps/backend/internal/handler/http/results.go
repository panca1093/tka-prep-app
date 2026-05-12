package http

import (
	"context"
	"errors"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	resultsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/result"
)

func (s *APIServer) GetResults(ctx context.Context, req api.GetResultsRequestObject) (api.GetResultsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetResults401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleStudent && claims.Role != domain.RoleAdmin {
		return api.GetResults403JSONResponse(errBody("FORBIDDEN", "student or admin only")), nil
	}

	results, err := s.resultSvc.List(ctx, claims.UserID, claims.Role)
	if err != nil {
		return nil, err
	}

	resp := api.GetResults200JSONResponse{Data: make([]api.TestResultResponse, 0, len(results))}
	for _, r := range results {
		resp.Data = append(resp.Data, toTestResultResponse(r))
	}
	return resp, nil
}

func (s *APIServer) GetResultsResultId(ctx context.Context, req api.GetResultsResultIdRequestObject) (api.GetResultsResultIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetResultsResultId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	detail, err := s.resultSvc.Get(ctx, req.ResultId, claims.UserID, claims.Role)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.GetResultsResultId404JSONResponse(errBody("NOT_FOUND", "result not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.GetResultsResultId403JSONResponse(errBody("FORBIDDEN", "not your result")), nil
		}
		return nil, err
	}

	return api.GetResultsResultId200JSONResponse(toResultDetailResponse(detail)), nil
}

func (s *APIServer) GetResultsResultIdReview(ctx context.Context, req api.GetResultsResultIdReviewRequestObject) (api.GetResultsResultIdReviewResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetResultsResultIdReview401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	filter := resultsvc.ReviewFilterAll
	if req.Params.Status != nil {
		filter = resultsvc.ReviewFilter(*req.Params.Status)
	}

	items, err := s.resultSvc.GetReview(ctx, req.ResultId, claims.UserID, claims.Role, filter)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.GetResultsResultIdReview404JSONResponse(errBody("NOT_FOUND", "result not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.GetResultsResultIdReview403JSONResponse(errBody("FORBIDDEN", "not your result")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.GetResultsResultIdReview422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}

	resp := api.GetResultsResultIdReview200JSONResponse{Data: make([]api.ReviewItemResponse, 0, len(items))}
	for _, item := range items {
		resp.Data = append(resp.Data, toReviewItemResponse(item))
	}
	return resp, nil
}

func toResultDetailResponse(d *domain.ResultDetail) api.ResultDetailResponse {
	tbs := make([]api.TopicBreakdownResponse, len(d.TopicBreakdown))
	for i, tb := range d.TopicBreakdown {
		tbs[i] = api.TopicBreakdownResponse{
			TopicId:      tb.TopicID,
			TopicName:    tb.TopicName,
			Total:        tb.Total,
			CorrectCount: tb.CorrectCount,
			WrongCount:   tb.WrongCount,
			BlankCount:   tb.BlankCount,
		}
	}
	return api.ResultDetailResponse{
		Id:             d.ID,
		SessionId:      d.SessionID,
		StudentId:      d.StudentID,
		TestId:         d.TestID,
		TotalScore:     d.TotalScore,
		CorrectCount:   d.CorrectCount,
		WrongCount:     d.WrongCount,
		BlankCount:     d.BlankCount,
		CompletedAt:    d.CompletedAt,
		Percentage:     d.Percentage,
		TopicBreakdown: tbs,
	}
}

func toReviewItemResponse(item domain.ReviewItem) api.ReviewItemResponse {
	opts := make([]api.ReviewOptionResponse, len(item.Options))
	for i, o := range item.Options {
		opts[i] = api.ReviewOptionResponse{
			Id:        o.ID,
			Label:     o.Label,
			Text:      o.Text,
			IsCorrect: o.IsCorrect,
		}
	}
	return api.ReviewItemResponse{
		QuestionId:       item.QuestionID,
		OrderIndex:       item.OrderIndex,
		Text:             item.Text,
		Explanation:      item.Explanation,
		Difficulty:       api.ReviewItemResponseDifficulty(item.Difficulty),
		TopicId:          item.TopicID,
		TopicName:        item.TopicName,
		Options:          opts,
		SelectedOptionId: item.SelectedOptionID,
		CorrectOptionId:  item.CorrectOptionID,
		Status:           api.ReviewItemResponseStatus(item.Status),
	}
}
