package storage

import (
	"context"
	"io"
	"net/http"
)

// FileStorage abstracts file persistence so uploads can target local disk or GCS
// without the handler layer knowing which.
type FileStorage interface {
	// Save persists data read from r and returns the public URL path for the
	// stored file (e.g. "/uploads/<uuid>.png").
	Save(ctx context.Context, filename string, r io.Reader, contentType string) (urlPath string, err error)

	// Delete removes a previously stored file given its URL path.
	Delete(ctx context.Context, urlPath string) error

	// ServeHTTP streams the stored file to the client.
	ServeHTTP(w http.ResponseWriter, r *http.Request, urlPath string)
}

const MaxUploadSize = 2 << 20 // 2 MB — enforced by the handler, declared here for reference
