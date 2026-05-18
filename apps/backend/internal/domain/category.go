package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	Name      string
	TestCount int
	CreatedAt time.Time
}
