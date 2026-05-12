package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	pkgjwt "github.com/yourorg/tkaprep/apps/backend/internal/pkg/jwt"
)

const maxUploadSize = 5 << 20 // 5 MB

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// UploadHandler handles POST /api/v1/upload.
// It is registered as a raw chi route (not via the strict generated handler)
// because multipart form data doesn't fit cleanly in the generated interface.
func UploadHandler(uploadDir string) http.HandlerFunc {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create upload dir: %v", err))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := pkgjwt.FromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 5 MB)"})
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
		dst, err := os.Create(filepath.Join(uploadDir, filename))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save file"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not write file"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"url": "/uploads/" + filename})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
