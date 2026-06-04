package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGCSStorage_NewGCSStorage_RequiresBucket(t *testing.T) {
	_, err := NewGCSStorage(GCSConfig{Bucket: ""})
	if err == nil {
		t.Error("expected error for empty bucket, got nil")
	}
}

func TestGCSStorage_Save_URLFormat(t *testing.T) {
	// Integration test: verifies the URL format produced by Save.
	// Requires a real GCS bucket — skip if not configured.
	t.Skip("requires GCS bucket — set GCS_BUCKET and credentials to run")

	ctx := context.Background()
	s, err := NewGCSStorage(GCSConfig{Bucket: "test-bucket"})
	if err != nil {
		t.Fatalf("NewGCSStorage: %v", err)
	}
	defer s.Close()

	urlPath, err := s.Save(ctx, "test-file.png", strings.NewReader("hello"), "image/png")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !strings.HasPrefix(urlPath, "/uploads/questions/") {
		t.Errorf("expected /uploads/questions/ prefix, got %s", urlPath)
	}
	if !strings.HasSuffix(urlPath, ".png") {
		t.Errorf("expected .png suffix, got %s", urlPath)
	}
}

func TestGCSStorage_Delete_ValidatesPath(t *testing.T) {
	s := &GCSStorage{bucket: "test"}
	ctx := context.Background()

	// Path without /uploads/ prefix should error.
	if err := s.Delete(ctx, "questions/foo.png"); err == nil {
		t.Error("expected error for path without /uploads/ prefix")
	}
}

func TestGCSStorage_ServeHTTP_ValidatesPrefix(t *testing.T) {
	s := &GCSStorage{bucket: "test", client: nil}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Without /uploads/ prefix — should return 400 before hitting GCS.
	// Will panic on nil client after the prefix check passes, so we test
	// only the path that fails the prefix check.
	s.ServeHTTP(rec, req, "/bad-path")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad path, got %d", rec.Code)
	}
}

func TestGCSStorage_ListByPrefix_URLFormat(t *testing.T) {
	t.Skip("requires GCS bucket — set GCS_BUCKET and credentials to run")

	ctx := context.Background()
	s, err := NewGCSStorage(GCSConfig{Bucket: "test-bucket"})
	if err != nil {
		t.Fatalf("NewGCSStorage: %v", err)
	}
	defer s.Close()

	paths, err := s.ListByPrefix(ctx, "questions/")
	if err != nil {
		t.Fatalf("ListByPrefix: %v", err)
	}
	for _, p := range paths {
		if !strings.HasPrefix(p, "/uploads/") {
			t.Errorf("expected /uploads/ prefix, got %s", p)
		}
	}
}
