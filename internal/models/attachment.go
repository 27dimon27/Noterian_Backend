package models

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID        uuid.UUID
	BlockID   uuid.UUID
	FileName  string
	FileSize  int64
	MimeType  string
	MinioKey  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
