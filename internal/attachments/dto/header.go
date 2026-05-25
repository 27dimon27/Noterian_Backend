package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Header struct {
	ID           uuid.UUID `json:"id"`
	NoteID       uuid.UUID `json:"note_id"`
	MinioKey     string    `json:"minio_key"`
	HeaderURL    string    `json:"header_url"`
	URLExpiresAt time.Time `json:"url_expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToHeaderDTO(header models.Header) Header {
	return Header{
		ID:           header.ID,
		NoteID:       header.NoteID,
		MinioKey:     header.MinioKey,
		HeaderURL:    header.HeaderURL,
		URLExpiresAt: header.URLExpiresAt,
		CreatedAt:    header.CreatedAt,
		UpdatedAt:    header.UpdatedAt,
	}
}
