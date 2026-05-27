package http

import (
	"context"
	"errors"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
)

func (s *APIServer) ListCategories(ctx context.Context, _ api.ListCategoriesRequestObject) (api.ListCategoriesResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.ListCategories401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	_ = claims // any authenticated user can list categories (needed for test builder dropdown)

	cats, err := s.categorySvc.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := api.ListCategories200JSONResponse{Data: make([]api.Category, 0, len(cats))}
	for _, c := range cats {
		resp.Data = append(resp.Data, domainCategoryToAPI(c))
	}
	return resp, nil
}

func (s *APIServer) CreateCategory(ctx context.Context, req api.CreateCategoryRequestObject) (api.CreateCategoryResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.CreateCategory401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.CreateCategory403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	cat, err := s.categorySvc.Create(ctx, req.Body.Name)
	if err != nil {
		if errors.Is(err, apierr.ErrConflict) || errors.Is(err, apierr.ErrValidation) {
			return api.CreateCategory400JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.CreateCategory201JSONResponse(domainCategoryToAPI(cat)), nil
}

func (s *APIServer) UpdateCategory(ctx context.Context, req api.UpdateCategoryRequestObject) (api.UpdateCategoryResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.UpdateCategory401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.UpdateCategory403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	uid := uuid.UUID(req.Id)
	name := ""
	if req.Body.Name != nil {
		name = *req.Body.Name
	}

	cat, err := s.categorySvc.Update(ctx, uid, name)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.UpdateCategory404JSONResponse(errBody("NOT_FOUND", "kategori tidak ditemukan")), nil
		case errors.Is(err, apierr.ErrConflict) || errors.Is(err, apierr.ErrValidation):
			return api.UpdateCategory400JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.UpdateCategory200JSONResponse(domainCategoryToAPI(cat)), nil
}

func (s *APIServer) UpdateOwnedCategory(ctx context.Context, req api.UpdateOwnedCategoryRequestObject) (api.UpdateOwnedCategoryResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.UpdateOwnedCategory401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleContributor {
		return api.UpdateOwnedCategory403JSONResponse(errBody("FORBIDDEN", "contributor only")), nil
	}

	callerID := uuid.UUID(claims.UserID)
	categoryID := uuid.UUID(req.Id)

	cat, err := s.categorySvc.UpdateOwned(ctx, callerID, categoryID, req.Body.Name, req.Body.Description)
	if err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.UpdateOwnedCategory404JSONResponse(errBody("NOT_FOUND", "kategori tidak ditemukan")), nil
		case errors.Is(err, apierr.ErrForbidden):
			return api.UpdateOwnedCategory403JSONResponse(errBody("FORBIDDEN", "bukan pemilik kategori")), nil
		case errors.Is(err, apierr.ErrValidation):
			return api.UpdateOwnedCategory400JSONResponse(errBody("VALIDATION_ERROR", err.Error())), nil
		}
		return nil, err
	}
	return api.UpdateOwnedCategory200JSONResponse(domainCategoryToAPI(cat)), nil
}

func (s *APIServer) DeleteCategory(ctx context.Context, req api.DeleteCategoryRequestObject) (api.DeleteCategoryResponseObject, error) {
	claims, ok := pkgjwt.FromContext(ctx)
	if !ok {
		return api.DeleteCategory401JSONResponse(errBody("UNAUTHORIZED", "not authenticated")), nil
	}
	if claims.Role != domain.RoleAdmin {
		return api.DeleteCategory403JSONResponse(errBody("FORBIDDEN", "admin only")), nil
	}

	uid := uuid.UUID(req.Id)

	if err := s.categorySvc.Delete(ctx, uid); err != nil {
		switch {
		case errors.Is(err, apierr.ErrNotFound):
			return api.DeleteCategory404JSONResponse(errBody("NOT_FOUND", "kategori tidak ditemukan")), nil
		case errors.Is(err, apierr.ErrConflict):
			return api.DeleteCategory409JSONResponse(errBody("CONFLICT", err.Error())), nil
		}
		return nil, err
	}
	return api.DeleteCategory200JSONResponse(api.MessageResponse{Message: "kategori dihapus"}), nil
}

func domainCategoryToAPI(c *domain.Category) api.Category {
	var createdBy *openapi_types.UUID
	if c.CreatedBy != nil {
		uid := openapi_types.UUID(*c.CreatedBy)
		createdBy = &uid
	}
	return api.Category{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedBy:   createdBy,
		TestCount:   c.TestCount,
		CreatedAt:   c.CreatedAt,
	}
}
