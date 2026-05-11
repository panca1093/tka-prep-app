package http

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler serves the /health endpoint defined in openapi.yaml.
// When the OpenAPI generation pipeline is wired up, this handler will
// implement the generated interface instead of being a free-standing struct.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   "0.1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
