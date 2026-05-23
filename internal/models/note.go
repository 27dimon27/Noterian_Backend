package models

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model Note
type Note struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Title      string     `json:"title"`
	ParentID   *uuid.UUID `json:"parent_id"`
	IsPublic   bool       `json:"is_public"`
	IsFavorite bool       `json:"is_favorite"`
	Icon       string     `json:"icon"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
