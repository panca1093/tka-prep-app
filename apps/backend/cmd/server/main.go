package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/yourorg/tkaprep/apps/backend/internal/config"
	httphandler "github.com/yourorg/tkaprep/apps/backend/internal/handler/http"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/server"
	pgstore "github.com/yourorg/tkaprep/apps/backend/internal/repository/postgres"
	adminsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/admin"
	authsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/auth"
	leadersvc "github.com/yourorg/tkaprep/apps/backend/internal/service/leaderboard"
	questionsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/question"
	resultsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/result"
	sessionsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/session"
	testsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/test"
	topicsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/topic"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Str("env", cfg.Env).Str("port", cfg.Port).Msg("starting tkaprep backend")

	ctx := context.Background()

	pool, err := pgstore.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	userRepo := pgstore.NewUserRepository(pool)
	tokenRepo := pgstore.NewRefreshTokenRepository(pool)
	topicRepo := pgstore.NewTopicRepository(pool)
	questionRepo := pgstore.NewQuestionRepository(pool)
	testRepo := pgstore.NewTestRepository(pool)
	sessionRepo := pgstore.NewSessionRepository(pool)
	resultRepo := pgstore.NewResultRepository(pool)
	leaderboardRepo := pgstore.NewLeaderboardRepository(pool)
	leaderboardSvc := leadersvc.NewService(leaderboardRepo)
	adminRepo := pgstore.NewAdminRepository(pool)

	authSvc := authsvc.NewService(authsvc.Config{
		JWTSecret:     cfg.JWTSecret,
		JWTAccessTTL:  cfg.JWTAccessTTL,
		JWTRefreshTTL: cfg.JWTRefreshTTL,
	}, userRepo, tokenRepo)
	topicService := topicsvc.NewService(topicRepo)
	questionService := questionsvc.NewService(questionRepo, cfg.UploadDir)
	testService := testsvc.NewService(testRepo)
	sessionService := sessionsvc.NewService(sessionRepo, resultRepo, testRepo, pool)
	resultService := resultsvc.NewService(resultRepo, sessionRepo)
	adminService := adminsvc.New(userRepo, adminRepo)

	apiHandler := httphandler.NewAPIServer(authSvc, topicService, questionService, testService, sessionService, resultService, leaderboardSvc, adminService)
	srv := server.New(cfg, apiHandler)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server error")
		}
	}()
	log.Info().Str("addr", httpServer.Addr).Msg("listening")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info().Msg("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	log.Info().Msg("bye")
}
