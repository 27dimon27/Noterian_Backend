package models

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model Avatar
type Avatar struct {
	ID           uuid.UUID
	ProfileID    uuid.UUID
	MinioKey     string
	AvatarURL    string
	URLExpiresAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
