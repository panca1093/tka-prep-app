package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/storage"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Env            string
	Port           string
	LogLevel       string
	DatabaseURL    string
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	AllowedOrigins []string
	UploadDir      string
	StorageBackend string // "local" (default) or "gcs"
	GCSBucket      string
}

// Storage returns a FileStorage implementation based on StorageBackend.
// Defaults to local if unset; panics on misconfigured GCS so the app fails fast.
func (c *Config) Storage() storage.FileStorage {
	backend := c.StorageBackend
	if backend == "" {
		backend = "local"
	}
	switch backend {
	case "gcs":
		return c.gcsStorage()
	default:
		return c.localStorage()
	}
}

func (c *Config) localStorage() storage.FileStorage {
	s, err := storage.NewLocalStorage(c.UploadDir)
	if err != nil {
		panic(fmt.Sprintf("local storage: %v", err))
	}
	return s
}

func (c *Config) gcsStorage() storage.FileStorage {
	s, err := storage.NewGCSStorage(storage.GCSConfig{
		Bucket: c.GCSBucket,
	})
	if err != nil {
		panic(fmt.Sprintf("gcs storage: %v", err))
	}
	return s
}

// Load reads configuration from environment variables, applying sensible defaults
// for local development. Required secrets (e.g. JWT_SECRET in non-local envs)
// must be provided explicitly.
func Load() (*Config, error) {
	cfg := &Config{
		Env:            getEnv("APP_ENV", "local"),
		Port:           getEnv("PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AllowedOrigins: strings.Split(getEnv("FRONTEND_ORIGIN", "http://localhost:5173"), ","),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		StorageBackend: getEnv("STORAGE_BACKEND", "local"),
		GCSBucket:      getEnv("GCS_BUCKET", ""),
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWTAccessTTL = accessTTL

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	cfg.JWTRefreshTTL = refreshTTL

	// In local env, missing DATABASE_URL is fine (we boot without DB to support /health).
	// In any other env, fail fast.
	if cfg.Env != "local" {
		if cfg.DatabaseURL == "" {
			return nil, fmt.Errorf("DATABASE_URL is required in env=%s", cfg.Env)
		}
		if cfg.JWTSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET is required in env=%s", cfg.Env)
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
