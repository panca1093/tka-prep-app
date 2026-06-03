package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_Save(t *testing.T) {
	dir := t.TempDir()
	s, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	ctx := context.Background()
	urlPath, err := s.Save(ctx, "abc-123.png", strings.NewReader("image-data"), "image/png")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if urlPath != "/uploads/abc-123.png" {
		t.Errorf("expected /uploads/abc-123.png, got %s", urlPath)
	}

	// Verify file exists on disk.
	data, err := os.ReadFile(filepath.Join(dir, "abc-123.png"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(data) != "image-data" {
		t.Errorf("expected image-data, got %s", string(data))
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocalStorage(dir)
	ctx := context.Background()

	// Create a file first.
	if err := os.WriteFile(filepath.Join(dir, "test-file.jpg"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(ctx, "/uploads/test-file.jpg"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "test-file.jpg")); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}

	// Delete non-existent file should not error.
	if err := s.Delete(ctx, "/uploads/does-not-exist.jpg"); err != nil {
		t.Errorf("Delete non-existent: %v", err)
	}

	// Delete invalid filename should error.
	if err := s.Delete(ctx, "/uploads/malicious.exe"); err == nil {
		t.Error("expected error for invalid extension")
	}
}

func TestLocalStorage_ServeHTTP(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocalStorage(dir)

	if err := os.WriteFile(filepath.Join(dir, "hello.png"), []byte("png-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/uploads/hello.png", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req, "/uploads/hello.png")

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "png-content" {
		t.Errorf("expected png-content, got %s", rec.Body.String())
	}
}

func TestLocalStorage_ServeHTTP_RejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocalStorage(dir)

	req := httptest.NewRequest(http.MethodGet, "/uploads/../../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req, "/uploads/../../../etc/passwd")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for path traversal, got %d", rec.Code)
	}
}
