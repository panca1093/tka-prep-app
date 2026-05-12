package http

import (
	"context"
	"errors"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/topic"
)

func (s *APIServer) GetTopics(ctx context.Context, _ api.GetTopicsRequestObject) (api.GetTopicsResponseObject, error) {
	topics, err := s.topicSvc.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := api.GetTopics200JSONResponse{Data: make([]api.TopicResponse, 0, len(topics))}
	for _, t := range topics {
		resp.Data = append(resp.Data, toTopicResponse(t))
	}
	return resp, nil
}

func (s *APIServer) PostTopics(ctx context.Context, req api.PostTopicsRequestObject) (api.PostTopicsResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PostTopics401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.PostTopics403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	t, err := s.topicSvc.Create(ctx, topic.CreateInput{
		Name:        req.Body.Name,
		Description: req.Body.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrConflict):
			return api.PostTopics422JSONResponse(errBody("CONFLICT", "topic name already exists")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PostTopics422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}

	return api.PostTopics201JSONResponse(toTopicResponse(t)), nil
}

func (s *APIServer) PatchTopicsTopicId(ctx context.Context, req api.PatchTopicsTopicIdRequestObject) (api.PatchTopicsTopicIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PatchTopicsTopicId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.PatchTopicsTopicId403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	t, err := s.topicSvc.Update(ctx, req.TopicId, topic.UpdateInput{
		Name:        req.Body.Name,
		Description: req.Body.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.PatchTopicsTopicId404JSONResponse(errBody("NOT_FOUND", "topic not found")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.PatchTopicsTopicId422JSONResponse(errBody("CONFLICT", "topic name already exists")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PatchTopicsTopicId422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}

	return api.PatchTopicsTopicId200JSONResponse(toTopicResponse(t)), nil
}

func (s *APIServer) DeleteTopicsTopicId(ctx context.Context, req api.DeleteTopicsTopicIdRequestObject) (api.DeleteTopicsTopicIdResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.DeleteTopicsTopicId401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.DeleteTopicsTopicId403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	if err := s.topicSvc.Delete(ctx, req.TopicId); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.DeleteTopicsTopicId404JSONResponse(errBody("NOT_FOUND", "topic not found")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.DeleteTopicsTopicId409JSONResponse(errBody("CONFLICT", "topic is in use by questions")), nil
		}
		return nil, err
	}

	return api.DeleteTopicsTopicId200JSONResponse{Message: "deleted"}, nil
}

func toTopicResponse(t *domain.Topic) api.TopicResponse {
	return api.TopicResponse{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
	}
}
