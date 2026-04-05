package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Attachment struct {
	ID        uuid.UUID `json:"id"`
	BlockID   uuid.UUID `json:"block_id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	MinioKey  string    `json:"minio_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToAttachmentDTO(attachment models.Attachment) Attachment {
	return Attachment{
		ID:        attachment.ID,
		BlockID:   attachment.BlockID,
		FileName:  attachment.FileName,
		FileSize:  attachment.FileSize,
		MimeType:  attachment.MimeType,
		MinioKey:  attachment.MinioKey,
		CreatedAt: attachment.CreatedAt,
		UpdatedAt: attachment.UpdatedAt,
	}
}
