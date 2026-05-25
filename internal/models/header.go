package models

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model Header
type Header struct {
	ID           uuid.UUID
	NoteID       uuid.UUID
	MinioKey     string
	HeaderURL    string
	URLExpiresAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
