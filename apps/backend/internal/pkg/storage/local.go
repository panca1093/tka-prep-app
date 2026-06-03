package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage stores files on the local filesystem — the current default.
type LocalStorage struct {
	uploadDir string
}

func NewLocalStorage(uploadDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir %s: %w", uploadDir, err)
	}
	return &LocalStorage{uploadDir: uploadDir}, nil
}

func (s *LocalStorage) Save(_ context.Context, filename string, r io.Reader, _ string) (string, error) {
	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return "/uploads/" + filename, nil
}

func (s *LocalStorage) Delete(_ context.Context, urlPath string) error {
	filename := filepath.Base(urlPath)
	if filename == "." || filename == "/" {
		return errors.New("invalid path")
	}
	// Only allow deleting files we'd have created (UUID-named with known extensions).
	if !strings.HasSuffix(filename, ".jpg") &&
		!strings.HasSuffix(filename, ".jpeg") &&
		!strings.HasSuffix(filename, ".png") &&
		!strings.HasSuffix(filename, ".gif") &&
		!strings.HasSuffix(filename, ".webp") {
		return errors.New("invalid filename")
	}
	if err := os.Remove(filepath.Join(s.uploadDir, filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}

func (s *LocalStorage) ServeHTTP(w http.ResponseWriter, r *http.Request, urlPath string) {
	// urlPath comes from the router, e.g. "/uploads/abc.jpg".
	// Strip "/uploads/" prefix if present.
	clean := strings.TrimPrefix(urlPath, "/uploads/")
	clean = filepath.Clean(clean)
	if strings.Contains(clean, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.uploadDir, clean))
}

// DirPath exposes the underlying upload directory path for use with
// http.FileServer (existing static route) and orphan cleanup. This is
// intentionally NOT on the FileStorage interface — it's local-only.
func (s *LocalStorage) DirPath() string {
	return s.uploadDir
}
