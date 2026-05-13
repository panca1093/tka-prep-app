package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
)

// publicPaths are exempt from JWT validation (exact match).
var publicPaths = map[string]bool{
	"/api/v1/health":        true,
	"/api/v1/auth/register": true,
	"/api/v1/auth/login":    true,
	"/api/v1/auth/refresh":  true,
}

func isPublic(path string) bool {
	if publicPaths[path] {
		return true
	}
	// Uploaded images are served publicly — browser img tags can't send Bearer tokens.
	return strings.HasPrefix(path, "/uploads/")
}

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublic(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeUnauthorized(w, "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := pkgjwt.ParseAccessToken(jwtSecret, tokenStr)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r.WithContext(pkgjwt.WithClaims(r.Context(), claims)))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := pkgjwt.FromContext(r.Context())
			if !ok {
				writeUnauthorized(w, "not authenticated")
				return
			}
			if !allowed[string(claims.Role)] {
				writeForbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", msg)
}

func writeForbidden(w http.ResponseWriter) {
	writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
