package http

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/auth"
)

func (s *APIServer) PostAuthRegister(ctx context.Context, req api.PostAuthRegisterRequestObject) (api.PostAuthRegisterResponseObject, error) {
	in := auth.RegisterInput{
		Name:     req.Body.Name,
		Email:    string(req.Body.Email),
		Password: req.Body.Password,
		Role:     domain.Role(req.Body.Role),
	}
	if req.Body.EducationLevel != nil {
		el := domain.EducationLevel(string(*req.Body.EducationLevel))
		in.EducationLevel = &el
	}

	user, pair, err := s.authSvc.Register(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrConflict):
			return api.PostAuthRegister409JSONResponse(errBody("EMAIL_CONFLICT", "email already registered")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.PostAuthRegister422JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}

	return api.PostAuthRegister201JSONResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toUserResponse(user),
	}, nil
}

func (s *APIServer) PostAuthLogin(ctx context.Context, req api.PostAuthLoginRequestObject) (api.PostAuthLoginResponseObject, error) {
	user, pair, err := s.authSvc.Login(ctx, string(req.Body.Email), req.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrUnauthorized):
			return api.PostAuthLogin401JSONResponse(errBody("UNAUTHORIZED", "invalid credentials")), nil
		case errors.Is(err, apierr.ErrPending):
			return api.PostAuthLogin401JSONResponse(errBody("PENDING_APPROVAL", "account is pending admin approval")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.PostAuthLogin401JSONResponse(errBody("SUSPENDED", "account suspended")), nil
		}
		return nil, err
	}

	return api.PostAuthLogin200JSONResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toUserResponse(user),
	}, nil
}

func (s *APIServer) PostAuthRefresh(ctx context.Context, req api.PostAuthRefreshRequestObject) (api.PostAuthRefreshResponseObject, error) {
	user, pair, err := s.authSvc.Refresh(ctx, req.Body.RefreshToken)
	if err != nil {
		if errors.Is(err, apierr.ErrUnauthorized) {
			return api.PostAuthRefresh401JSONResponse(errBody("UNAUTHORIZED", "invalid or expired refresh token")), nil
		}
		return nil, err
	}

	return api.PostAuthRefresh200JSONResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toUserResponse(user),
	}, nil
}

func (s *APIServer) PostAuthLogout(ctx context.Context, req api.PostAuthLogoutRequestObject) (api.PostAuthLogoutResponseObject, error) {
	if err := s.authSvc.Logout(ctx, req.Body.RefreshToken); err != nil {
		return nil, err
	}
	return api.PostAuthLogout200JSONResponse{Message: "logged out"}, nil
}

func (s *APIServer) GetAuthMe(ctx context.Context, _ api.GetAuthMeRequestObject) (api.GetAuthMeResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.GetAuthMe401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}

	user, err := s.authSvc.Me(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			return api.GetAuthMe401JSONResponse(errBody("UNAUTHORIZED", "user not found")), nil
		}
		return nil, err
	}

	return api.GetAuthMe200JSONResponse(toUserResponse(user)), nil
}

func (s *APIServer) PatchAuthMe(ctx context.Context, req api.PatchAuthMeRequestObject) (api.PatchAuthMeResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.PatchAuthMe401Response{}, nil
	}

	in := auth.UpdateProfileInput{}
	if req.Body.EducationLevel != nil {
		el := domain.EducationLevel(string(*req.Body.EducationLevel))
		in.EducationLevel = &el
	}
	if req.Body.Gender != nil {
		g := domain.Gender(string(*req.Body.Gender))
		in.Gender = &g
	}
	if req.Body.Phone != nil {
		in.Phone = req.Body.Phone
	}
	if req.Body.AvatarUrl != nil {
		in.AvatarURL = req.Body.AvatarUrl
	}

	user, err := s.authSvc.UpdateProfile(ctx, claims.UserID, in)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return nil, apierr.ErrNotFound
		case errors.Is(err, apierr.ErrValidation):
			return api.PatchAuthMe400Response{}, nil
		}
		return nil, err
	}

	return api.PatchAuthMe200JSONResponse(toUserResponse(user)), nil
}

// toUserResponse maps a domain User to the generated API type.
func toUserResponse(u *domain.User) api.UserResponse {
	resp := api.UserResponse{
		Id:        u.ID,
		Name:      u.Name,
		Email:     openapi_types.Email(u.Email),
		Role:      api.UserResponseRole(u.Role),
		Status:    api.UserResponseStatus(u.Status),
		CreatedAt: u.CreatedAt,
	}
	if u.EducationLevel != nil {
		el := api.UserResponseEducationLevel(string(*u.EducationLevel))
		resp.EducationLevel = &el
	}
	if u.Gender != nil {
		g := api.UserResponseGender(string(*u.Gender))
		resp.Gender = &g
	}
	if u.Phone != nil {
		resp.Phone = u.Phone
	}
	if u.AvatarURL != nil {
		resp.AvatarUrl = u.AvatarURL
	}
	return resp
}

// errBody builds a typed ErrorResponse value for use as a response body.
func errBody(code, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Details *[]struct {
				Field *string `json:"field,omitempty"`
				Issue *string `json:"issue,omitempty"`
			} `json:"details,omitempty"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}
}
