package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Attachment struct {
	ID           uuid.UUID `json:"id"`
	BlockID      uuid.UUID `json:"block_id"`
	MinioKey     string    `json:"minio_key"`
	AttachURL    string    `json:"attach_url"`
	URLExpiresAt time.Time `json:"url_expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToAttachmentDTO(attachment models.Attachment) Attachment {
	return Attachment{
		ID:           attachment.ID,
		BlockID:      attachment.BlockID,
		MinioKey:     attachment.MinioKey,
		AttachURL:    attachment.AttachURL,
		URLExpiresAt: attachment.URLExpiresAt,
		CreatedAt:    attachment.CreatedAt,
		UpdatedAt:    attachment.UpdatedAt,
	}
}
