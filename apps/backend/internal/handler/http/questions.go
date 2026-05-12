package http

import (
	"context"
	"errors"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/question"
)

func (s *APIServer) GetQuestions(ctx context.Context, req api.GetQuestionsRequestObject) (api.GetQuestionsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetQuestions401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.GetQuestions403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	p := req.Params
	f := question.ListFilter{Page: 1, Limit: 20}
	if p.Page != nil {
		f.Page = *p.Page
	}
	if p.Limit != nil {
		f.Limit = *p.Limit
	}
	if p.Search != nil {
		f.Search = *p.Search
	}
	if p.TopicId != nil {
		f.TopicID = p.TopicId
	}
	if p.Difficulty != nil {
		d := domain.Difficulty(*p.Difficulty)
		f.Difficulty = &d
	}

	qs, total, err := s.questionSvc.List(ctx, f)
	if err != nil {
		return nil, err
	}

	resp := api.GetQuestions200JSONResponse{
		Data:  make([]api.QuestionDetailResponse, 0, len(qs)),
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	}
	for _, q := range qs {
		resp.Data = append(resp.Data, toQuestionDetailResponse(q))
	}
	return resp, nil
}

func (s *APIServer) PostQuestions(ctx context.Context, req api.PostQuestionsRequestObject) (api.PostQuestionsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostQuestions401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PostQuestions403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	b := req.Body
	qt := domain.QuestionTypeMCQ
	if b.QuestionType != "" {
		qt = domain.QuestionType(b.QuestionType)
	}

	in := question.CreateInput{
		ContributorID: claims.UserID,
		TopicID:       b.TopicId,
		QuestionType:  qt,
		Text:          b.Text,
		Explanation:   b.Explanation,
		ImageURL:      b.ImageUrl,
		Difficulty:    domain.Difficulty(b.Difficulty),
	}
	if b.Options != nil {
		for _, o := range *b.Options {
			in.Options = append(in.Options, question.OptionInput{Label: o.Label, Text: o.Text, IsCorrect: o.IsCorrect, ImageURL: o.ImageUrl})
		}
	}
	if b.Statements != nil {
		for _, st := range *b.Statements {
			in.Statements = append(in.Statements, question.StatementInput{Text: st.Text, IsCorrect: st.IsCorrect, Position: st.Position, ImageURL: st.ImageUrl})
		}
	}

	q, err := s.questionSvc.Create(ctx, in)
	if err != nil {
		if errors.Is(err, apierr.ErrValidation) {
			return api.PostQuestions422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.PostQuestions201JSONResponse(toQuestionDetailResponse(q)), nil
}

func (s *APIServer) GetQuestionsQuestionId(ctx context.Context, req api.GetQuestionsQuestionIdRequestObject) (api.GetQuestionsQuestionIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetQuestionsQuestionId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.GetQuestionsQuestionId403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	q, err := s.questionSvc.Get(ctx, req.QuestionId)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return api.GetQuestionsQuestionId404JSONResponse(errBody("NOT_FOUND", "question not found")), nil
		}
		return nil, err
	}
	return api.GetQuestionsQuestionId200JSONResponse(toQuestionDetailResponse(q)), nil
}

func (s *APIServer) PatchQuestionsQuestionId(ctx context.Context, req api.PatchQuestionsQuestionIdRequestObject) (api.PatchQuestionsQuestionIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PatchQuestionsQuestionId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.PatchQuestionsQuestionId403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	b := req.Body
	in := question.UpdateInput{
		Text:        b.Text,
		Explanation: b.Explanation,
		ImageURL:    b.ImageUrl,
		TopicID:     b.TopicId,
	}
	if b.Difficulty != nil {
		d := domain.Difficulty(*b.Difficulty)
		in.Difficulty = &d
	}
	if b.Options != nil {
		for _, o := range *b.Options {
			in.Options = append(in.Options, question.OptionInput{Label: o.Label, Text: o.Text, IsCorrect: o.IsCorrect, ImageURL: o.ImageUrl})
		}
	}
	if b.Statements != nil {
		for _, st := range *b.Statements {
			in.Statements = append(in.Statements, question.StatementInput{Text: st.Text, IsCorrect: st.IsCorrect, Position: st.Position, ImageURL: st.ImageUrl})
		}
	}

	q, err := s.questionSvc.Update(ctx, req.QuestionId, claims.UserID, claims.Role, in)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PatchQuestionsQuestionId404JSONResponse(errBody("NOT_FOUND", "question not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PatchQuestionsQuestionId403JSONResponse(errBody("FORBIDDEN", "not your question")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PatchQuestionsQuestionId422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.PatchQuestionsQuestionId200JSONResponse(toQuestionDetailResponse(q)), nil
}

func (s *APIServer) DeleteQuestionsQuestionId(ctx context.Context, req api.DeleteQuestionsQuestionIdRequestObject) (api.DeleteQuestionsQuestionIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.DeleteQuestionsQuestionId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
		return api.DeleteQuestionsQuestionId403JSONResponse(errBody("FORBIDDEN", "contributor or admin only")), nil
	}

	if err := s.questionSvc.Delete(ctx, req.QuestionId, claims.UserID, claims.Role); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.DeleteQuestionsQuestionId404JSONResponse(errBody("NOT_FOUND", "question not found")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.DeleteQuestionsQuestionId403JSONResponse(errBody("FORBIDDEN", "not your question")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.DeleteQuestionsQuestionId409JSONResponse(errBody("CONFLICT", "question is used in a published test")), nil
		}
		return nil, err
	}
	return api.DeleteQuestionsQuestionId200JSONResponse{Message: "deleted"}, nil
}

func toQuestionDetailResponse(q *domain.Question) api.QuestionDetailResponse {
	opts := make([]api.QuestionOptionResponse, len(q.Options))
	for i, o := range q.Options {
		opts[i] = api.QuestionOptionResponse{Id: o.ID, Label: o.Label, Text: o.Text, IsCorrect: o.IsCorrect, ImageUrl: o.ImageURL}
	}
	stmts := make([]api.QuestionStatementResponse, len(q.Statements))
	for i, s := range q.Statements {
		stmts[i] = api.QuestionStatementResponse{Id: s.ID, Text: s.Text, IsCorrect: s.IsCorrect, Position: s.Position, ImageUrl: s.ImageURL}
	}
	return api.QuestionDetailResponse{
		Id:            q.ID,
		ContributorId: q.ContributorID,
		TopicId:       q.TopicID,
		QuestionType:  api.QuestionDetailResponseQuestionType(q.Type),
		Text:          q.Text,
		Explanation:   q.Explanation,
		ImageUrl:      q.ImageURL,
		Difficulty:    api.QuestionDetailResponseDifficulty(q.Difficulty),
		Options:       opts,
		Statements:    stmts,
		CreatedAt:     q.CreatedAt,
		UpdatedAt:     q.UpdatedAt,
	}
}
