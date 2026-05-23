package models

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model Profile
type Profile struct {
	ID           uuid.UUID
	Username     string
	Avatar       string
	Password     []byte
	TokenVersion int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
