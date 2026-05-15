package http

import (
	"context"
	"errors"
	"fmt"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/session"
)

func (s *APIServer) PostTestsTestIdSessions(ctx context.Context, req api.PostTestsTestIdSessionsRequestObject) (api.PostTestsTestIdSessionsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostTestsTestIdSessions401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleStudent {
		return api.PostTestsTestIdSessions403JSONResponse(errBody("FORBIDDEN", "student only")), nil
	}

	sess, err := s.sessionSvc.Start(ctx, claims.UserID, req.TestId)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PostTestsTestIdSessions404JSONResponse(errBody("NOT_FOUND", "test not found or not published")), nil
		case errors.Is(err, apierr.ErrConflict):
			return nil, fmt.Errorf("ALREADY_COMPLETED: %w", err)
		}
		return nil, err
	}
	return api.PostTestsTestIdSessions201JSONResponse(toSessionResponse(sess)), nil
}

func (s *APIServer) GetSessionsSessionId(ctx context.Context, req api.GetSessionsSessionIdRequestObject) (api.GetSessionsSessionIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetSessionsSessionId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	sess, err := s.sessionSvc.Get(ctx, req.SessionId, claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.GetSessionsSessionId404JSONResponse(errBody("NOT_FOUND", "session not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.GetSessionsSessionId403JSONResponse(errBody("FORBIDDEN", "not your session")), nil
		}
		return nil, err
	}
	return api.GetSessionsSessionId200JSONResponse(toSessionResponse(sess)), nil
}

func (s *APIServer) PostSessionsSessionIdAnswers(ctx context.Context, req api.PostSessionsSessionIdAnswersRequestObject) (api.PostSessionsSessionIdAnswersResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostSessionsSessionIdAnswers401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	b := req.Body
	in := session.AnswerInput{QuestionID: b.QuestionId}

	// MCQ
	in.SelectedOptionID = b.SelectedOptionId

	// PGK
	if b.SelectedOptionIds != nil {
		in.SelectedOptionIDs = *b.SelectedOptionIds
	}

	// B/S
	if b.StatementAnswers != nil {
		for _, sa := range *b.StatementAnswers {
			in.StatementAnswers = append(in.StatementAnswers, domain.StatementAnswerInput{
				StatementID: sa.StatementId,
				IsCorrect:   sa.IsCorrect,
			})
		}
	}

	sess, err := s.sessionSvc.SaveAnswer(ctx, req.SessionId, claims.UserID, in)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PostSessionsSessionIdAnswers404JSONResponse(errBody("NOT_FOUND", "session not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PostSessionsSessionIdAnswers403JSONResponse(errBody("FORBIDDEN", "not your session")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PostSessionsSessionIdAnswers409JSONResponse(errBody("CONFLICT", err.Error())), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PostSessionsSessionIdAnswers422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.PostSessionsSessionIdAnswers200JSONResponse(toSessionResponse(sess)), nil
}

func (s *APIServer) PostSessionsSessionIdFlag(ctx context.Context, req api.PostSessionsSessionIdFlagRequestObject) (api.PostSessionsSessionIdFlagResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostSessionsSessionIdFlag401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	sess, err := s.sessionSvc.ToggleFlag(ctx, req.SessionId, claims.UserID, req.Body.QuestionId, req.Body.IsFlagged)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PostSessionsSessionIdFlag404JSONResponse(errBody("NOT_FOUND", "session not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PostSessionsSessionIdFlag403JSONResponse(errBody("FORBIDDEN", "not your session")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PostSessionsSessionIdFlag409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.PostSessionsSessionIdFlag200JSONResponse(toSessionResponse(sess)), nil
}

func (s *APIServer) PostSessionsSessionIdSubmit(ctx context.Context, req api.PostSessionsSessionIdSubmitRequestObject) (api.PostSessionsSessionIdSubmitResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostSessionsSessionIdSubmit401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	result, err := s.sessionSvc.Submit(ctx, req.SessionId, claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PostSessionsSessionIdSubmit404JSONResponse(errBody("NOT_FOUND", "session not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PostSessionsSessionIdSubmit403JSONResponse(errBody("FORBIDDEN", "not your session")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PostSessionsSessionIdSubmit409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.PostSessionsSessionIdSubmit200JSONResponse(toTestResultResponse(result)), nil
}

func toSessionResponse(s *domain.TestSession) api.SessionResponse {
	answers := make([]api.SessionAnswerResponse, len(s.Answers))
	for i, a := range s.Answers {
		answers[i] = api.SessionAnswerResponse{
			Id:               a.ID,
			SessionId:        a.SessionID,
			QuestionId:       a.QuestionID,
			SelectedOptionId: a.SelectedOptionID,
			StatementId:      a.StatementID,
			BooleanAnswer:    a.BooleanAnswer,
			AnsweredAt:       a.AnsweredAt,
		}
	}
	flags := make([]api.SessionQuestionFlagResponse, len(s.Flags))
	for i, f := range s.Flags {
		flags[i] = api.SessionQuestionFlagResponse{
			QuestionId: f.QuestionID,
			IsFlagged:  f.IsFlagged,
		}
	}
	return api.SessionResponse{
		Id:                   s.ID,
		StudentId:            s.StudentID,
		TestId:               s.TestID,
		StartedAt:            s.StartedAt,
		SubmittedAt:          s.SubmittedAt,
		Status:               api.SessionResponseStatus(s.Status),
		TimeRemainingSeconds: s.TimeRemainingSeconds,
		Answers:              answers,
		Flags:                &flags,
	}
}

func toTestResultResponse(r *domain.TestResult) api.TestResultResponse {
	return api.TestResultResponse{
		Id:           r.ID,
		SessionId:    r.SessionID,
		StudentId:    r.StudentID,
		TestId:       r.TestID,
		TestTitle:    r.TestTitle,
		TotalScore:   r.TotalScore,
		CorrectCount: r.CorrectCount,
		WrongCount:   r.WrongCount,
		BlankCount:   r.BlankCount,
		CompletedAt:  r.CompletedAt,
	}
}
