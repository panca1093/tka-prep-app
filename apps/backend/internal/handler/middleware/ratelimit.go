package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
)

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // requests per window
	window   time.Duration
	maxSlots int
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

func newRateLimiter(rate int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		window:   window,
		maxSlots: rate,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.Sub(b.lastSeen) > rl.window {
		rl.buckets[key] = &bucket{tokens: rl.maxSlots - 1, lastSeen: now}
		return true
	}

	b.lastSeen = now
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// clean up stale buckets every 5 minutes.
func (rl *rateLimiter) startCleanup(interval time.Duration) {
	go func() {
		for range time.NewTicker(interval).C {
			rl.mu.Lock()
			now := time.Now()
			for k, b := range rl.buckets {
				if now.Sub(b.lastSeen) > rl.window*2 {
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

var (
	authLimiter    = newRateLimiter(5, time.Minute)     // 5 req/min
	generalLimiter = newRateLimiter(100, time.Minute)   // 100 req/min
)

func init() {
	authLimiter.startCleanup(5 * time.Minute)
	generalLimiter.startCleanup(5 * time.Minute)
}

// authPaths are rate-limited at 5 req/min per IP.
var authPaths = map[string]bool{
	"/api/v1/auth/register": true,
	"/api/v1/auth/login":    true,
	"/api/v1/auth/refresh":  true,
}

func isAuthPath(path string) bool {
	return authPaths[path]
}

// RateLimit applies per-minute rate limiting:
//   - Auth endpoints: 5 requests/minute per IP
//   - All other authenticated endpoints: 100 requests/minute per user
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if isAuthPath(path) {
			ip := clientIP(r)
			if !authLimiter.allow(ip) {
				writeRateLimited(w)
				return
			}
		} else if !isPublic(path) {
			// General rate limit by user ID (or IP if unauthenticated).
			key := clientIP(r)
			if claims, ok := pkgjwt.FromContext(r.Context()); ok {
				key = claims.UserID.String()
			}
			if !generalLimiter.allow(key) {
				writeRateLimited(w)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	// Strip port from RemoteAddr.
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

func writeRateLimited(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "RATE_LIMITED",
			"message": "too many requests — slow down",
		},
	})
}
