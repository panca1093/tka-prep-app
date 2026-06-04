package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSConfig holds GCS-specific configuration.
type GCSConfig struct {
	Bucket string
}

// NewGCSStorage creates a GCS-backed FileStorage. Returns an error if the
// client cannot be initialized (bad credentials, missing bucket, etc.).
func NewGCSStorage(cfg GCSConfig) (*GCSStorage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("GCS_BUCKET is required")
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GCS client: %w", err)
	}
	// Verify the bucket exists and is accessible.
	if _, err := client.Bucket(cfg.Bucket).Attrs(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("access GCS bucket %s: %w", cfg.Bucket, err)
	}
	return &GCSStorage{
		bucket: cfg.Bucket,
		client: client,
	}, nil
}

// GCSStorage stores files in Google Cloud Storage.
type GCSStorage struct {
	bucket string
	client *storage.Client
}

// Save writes data to GCS under "questions/<filename>" and returns the URL path
// that the proxy handler expects.
func (s *GCSStorage) Save(ctx context.Context, filename string, r io.Reader, contentType string) (string, error) {
	// Store under a prefix to keep the bucket organized.
	objectPath := "questions/" + filename

	w := s.client.Bucket(s.bucket).Object(objectPath).NewWriter(ctx)
	w.ContentType = contentType
	// Cache publicly readable for 1 year (images are immutable — edits create new uploads).
	w.CacheControl = "public, max-age=31536000"

	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return "", fmt.Errorf("write to GCS: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("close GCS writer: %w", err)
	}

	// Return the proxy URL path — clients never see the GCS URL directly.
	return "/uploads/" + objectPath, nil
}

// Delete removes a GCS object. urlPath is expected to be "/uploads/questions/<filename>".
func (s *GCSStorage) Delete(ctx context.Context, urlPath string) error {
	objectPath := strings.TrimPrefix(urlPath, "/uploads/")
	if objectPath == "" || objectPath == urlPath {
		return fmt.Errorf("invalid url path: %s", urlPath)
	}
	if err := s.client.Bucket(s.bucket).Object(objectPath).Delete(ctx); err != nil {
		// Don't treat not-found as an error — consistent with local storage behavior.
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("delete gcs object %s: %w", objectPath, err)
	}
	return nil
}

// ServeHTTP proxies the request to GCS, streaming the object to the client.
func (s *GCSStorage) ServeHTTP(w http.ResponseWriter, r *http.Request, urlPath string) {
	objectPath := strings.TrimPrefix(urlPath, "/uploads/")
	if objectPath == "" || objectPath == urlPath {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	obj := s.client.Bucket(s.bucket).Object(objectPath)
	attrs, err := obj.Attrs(r.Context())
	if err != nil {
		if err == storage.ErrObjectNotExist {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rc, err := obj.NewReader(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	if attrs.ContentType != "" {
		w.Header().Set("Content-Type", attrs.ContentType)
	}
	if attrs.CacheControl != "" {
		w.Header().Set("Cache-Control", attrs.CacheControl)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attrs.Size))
	if _, err := io.Copy(w, rc); err != nil {
		// Log but can't change status — headers already written by io.Copy triggering WriteHeader.
		return
	}
}

// Close releases the GCS client connection. Call during graceful shutdown.
func (s *GCSStorage) Close() error {
	return s.client.Close()
}

// ListByPrefix returns object paths under a given prefix, for orphan cleanup.
// Used internally by the question service to find all uploaded images.
func (s *GCSStorage) ListByPrefix(ctx context.Context, prefix string) ([]string, error) {
	it := s.client.Bucket(s.bucket).Objects(ctx, &storage.Query{Prefix: prefix})
	var paths []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list gcs objects: %w", err)
		}
		paths = append(paths, "/uploads/"+attrs.Name)
	}
	return paths, nil
}
