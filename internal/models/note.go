package models

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model Note
type Note struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	ParentID  *uuid.UUID
	IsPublic  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
