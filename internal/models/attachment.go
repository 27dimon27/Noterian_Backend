package models

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID           uuid.UUID
	BlockID      uuid.UUID
	MinioKey     string
	AttachURL    string
	URLExpiresAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
