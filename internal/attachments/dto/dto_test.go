package dto

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

func TestToAttachmentDTO(t *testing.T) {
	attachmentID := uuid.New()
	blockID := uuid.New()
	now := time.Now()

	attachment := models.Attachment{
		ID:           attachmentID,
		BlockID:      blockID,
		MinioKey:     "test-key",
		AttachURL:    "https://example.com/file",
		URLExpiresAt: now.Add(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	dtoAttachment := ToAttachmentDTO(attachment)

	if dtoAttachment.ID != attachmentID {
		t.Errorf("expected ID %v, got %v", attachmentID, dtoAttachment.ID)
	}
	if dtoAttachment.BlockID != blockID {
		t.Errorf("expected BlockID %v, got %v", blockID, dtoAttachment.BlockID)
	}
	if dtoAttachment.MinioKey != "test-key" {
		t.Errorf("expected MinioKey 'test-key', got '%s'", dtoAttachment.MinioKey)
	}
	if dtoAttachment.AttachURL != "https://example.com/file" {
		t.Errorf("expected AttachURL 'https://example.com/file', got '%s'", dtoAttachment.AttachURL)
	}
}
