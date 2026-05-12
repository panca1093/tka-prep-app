package http

import (
	"context"
	"time"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
)

func (s *APIServer) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{
		Status:    api.Ok,
		Timestamp: time.Now().UTC(),
		Version:   strPtr("0.1.0"),
	}, nil
}

func strPtr(s string) *string { return &s }
