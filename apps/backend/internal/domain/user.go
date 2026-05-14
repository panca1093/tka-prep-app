package domain

import (
	"time"

	"github.com/google/uuid"
)

type EducationLevel string

const (
	EducationLevelSD  EducationLevel = "sd"
	EducationLevelSMP EducationLevel = "smp"
	EducationLevelSMA EducationLevel = "sma"
	EducationLevelSMK EducationLevel = "smk"
)

type Role string

const (
	RoleStudent     Role = "student"
	RoleContributor Role = "contributor"
	RoleAdmin       Role = "admin"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
	StatusPending   Status = "pending"
)

type User struct {
	ID             uuid.UUID
	Name           string
	Email          string
	PasswordHash   string
	Role           Role
	Status         Status
	EducationLevel *EducationLevel
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
