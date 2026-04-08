package models

import (
	"time"

	"github.com/google/uuid"
)

type Block struct {
	ID          uuid.UUID
	NoteID      uuid.UUID
	BlockTypeID int
	Position    int
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
