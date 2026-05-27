package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedBy   *uuid.UUID
	TestCount   int
	CreatedAt   time.Time
}
