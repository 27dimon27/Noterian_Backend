package models

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model Block
type Block struct {
	ID          uuid.UUID `json:"id"`
	NoteID      uuid.UUID `json:"note_id"`
	BlockTypeID int       `json:"block_type_id"`
	Position    int       `json:"position"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
