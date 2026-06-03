package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/api"
	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/storage"
)

const maxUploadSize = storage.MaxUploadSize

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// UploadHandler handles POST /api/v1/upload.
// It is registered as a raw chi route (not via the strict generated handler)
// because multipart form data doesn't fit cleanly in the generated interface.
func UploadHandler(store storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := pkgjwt.FromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}
		if claims.Role != domain.RoleContributor && claims.Role != domain.RoleAdmin {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "contributor role required"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 2 MB)"})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file field"})
			return
		}
		defer file.Close()

		// Detect MIME from first 512 bytes.
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mime := http.DetectContentType(buf[:n])
		ext, ok := allowedMIME[mime]
		if !ok {
			// Fall back to extension check for some formats DetectContentType misses.
			ext = strings.ToLower(filepath.Ext(header.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported file type"})
				return
			}
			if ext == ".jpeg" {
				ext = ".jpg"
			}
		}

		// Reset reader to start.
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		filename := uuid.New().String() + ext
		urlPath, err := store.Save(r.Context(), filename, file, mime)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save file"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"url": urlPath})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// UploadFile satisfies api.StrictServerInterface. The actual upload is handled
// by the raw chi route registered in server.go, which takes precedence over
// the strict handler mount, so this method is never invoked at runtime.
func (s *APIServer) UploadFile(_ context.Context, _ api.UploadFileRequestObject) (api.UploadFileResponseObject, error) {
	return api.UploadFile200JSONResponse{Url: ""}, nil
}
