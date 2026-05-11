package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/yourorg/tkaprep/apps/backend/internal/config"
	httphandler "github.com/yourorg/tkaprep/apps/backend/internal/handler/http"
)

// Server is the top-level HTTP server. It owns the router and any dependencies
// that need to be passed into handlers.
type Server struct {
	cfg    *config.Config
	router chi.Router
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:    cfg,
		router: chi.NewRouter(),
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	health := httphandler.NewHealthHandler()

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", health.Get)
	})

	// Outside /api/v1 too, for the simplest "is it up" probe
	s.router.Get("/health", health.Get)
}
