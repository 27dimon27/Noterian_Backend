package models

import (
	"time"

	"github.com/google/uuid"
)

type BlockState struct {
	ID         uuid.UUID
	BlockID    uuid.UUID
	Formatting map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
