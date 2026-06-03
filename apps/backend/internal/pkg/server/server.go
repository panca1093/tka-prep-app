package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/config"
	httphandler "github.com/yourorg/tkaprep/apps/backend/internal/handler/http"
	"github.com/yourorg/tkaprep/apps/backend/internal/handler/middleware"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/storage"
)

type Server struct {
	router chi.Router
}

// storageAdapter bridges the FileStorage interface to the chi static file
// handler or a GCS proxy, depending on the backend.
type storageAdapter struct {
	store storage.FileStorage
}

func (a *storageAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.store.ServeHTTP(w, r, r.URL.Path)
}

func New(cfg *config.Config, handler *httphandler.APIServer) *Server {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.Authenticate(cfg.JWTSecret))
	r.Use(middleware.RateLimit)

	// Bootstrap storage backend.
	store := cfg.Storage()

	// Upload endpoint (raw route — multipart doesn't fit strict handler pattern).
	r.Post("/api/v1/upload", httphandler.UploadHandler(store))

	// Static file serving for uploaded images.
	r.Handle("/uploads/*", &storageAdapter{store: store})

	strict := api.NewStrictHandler(handler, nil)
	r.Mount("/api/v1", api.HandlerFromMux(strict, chi.NewRouter()))

	return &Server{router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
