package apierr

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrPending      = errors.New("pending approval")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation error")
)
