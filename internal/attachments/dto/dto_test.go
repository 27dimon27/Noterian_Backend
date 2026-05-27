package dto

import (
	"encoding/json"
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
	if !dtoAttachment.URLExpiresAt.Equal(attachment.URLExpiresAt) {
		t.Errorf("expected URLExpiresAt %v, got %v", attachment.URLExpiresAt, dtoAttachment.URLExpiresAt)
	}
	if !dtoAttachment.CreatedAt.Equal(attachment.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", attachment.CreatedAt, dtoAttachment.CreatedAt)
	}
	if !dtoAttachment.UpdatedAt.Equal(attachment.UpdatedAt) {
		t.Errorf("expected UpdatedAt %v, got %v", attachment.UpdatedAt, dtoAttachment.UpdatedAt)
	}
}

func TestToHeaderDTO(t *testing.T) {
	headerID := uuid.New()
	noteID := uuid.New()
	now := time.Now()

	header := models.Header{
		ID:           headerID,
		NoteID:       noteID,
		MinioKey:     "header-key",
		HeaderURL:    "https://example.com/header",
		URLExpiresAt: now.Add(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	dtoHeader := ToHeaderDTO(header)

	if dtoHeader.ID != headerID {
		t.Errorf("expected ID %v, got %v", headerID, dtoHeader.ID)
	}
	if dtoHeader.NoteID != noteID {
		t.Errorf("expected NoteID %v, got %v", noteID, dtoHeader.NoteID)
	}
	if dtoHeader.MinioKey != "header-key" {
		t.Errorf("expected MinioKey 'header-key', got '%s'", dtoHeader.MinioKey)
	}
	if dtoHeader.HeaderURL != "https://example.com/header" {
		t.Errorf("expected HeaderURL 'https://example.com/header', got '%s'", dtoHeader.HeaderURL)
	}
	if !dtoHeader.URLExpiresAt.Equal(header.URLExpiresAt) {
		t.Errorf("expected URLExpiresAt %v, got %v", header.URLExpiresAt, dtoHeader.URLExpiresAt)
	}
	if !dtoHeader.CreatedAt.Equal(header.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", header.CreatedAt, dtoHeader.CreatedAt)
	}
	if !dtoHeader.UpdatedAt.Equal(header.UpdatedAt) {
		t.Errorf("expected UpdatedAt %v, got %v", header.UpdatedAt, dtoHeader.UpdatedAt)
	}
}

func TestAttachmentJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := Attachment{
		ID:           uuid.New(),
		BlockID:      uuid.New(),
		MinioKey:     "k",
		AttachURL:    "https://example.com/a",
		URLExpiresAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Attachment
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != original.ID || got.BlockID != original.BlockID || got.AttachURL != original.AttachURL {
		t.Fatalf("round trip mismatch: %+v vs %+v", got, original)
	}
}

func TestHeaderJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := Header{
		ID:           uuid.New(),
		NoteID:       uuid.New(),
		MinioKey:     "k",
		HeaderURL:    "https://example.com/h",
		URLExpiresAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Header
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != original.ID || got.NoteID != original.NoteID || got.HeaderURL != original.HeaderURL {
		t.Fatalf("round trip mismatch: %+v vs %+v", got, original)
	}
}
