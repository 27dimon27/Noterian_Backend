package websocket

import (
	"bytes"
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type AttachmentUsecaseAdapter struct {
	uploadFunc func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader, hasPosition bool, position int) (*models.Attachment, error)
}

func NewAttachmentUsecaseAdapter(
	uploadFunc func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader, hasPosition bool, position int) (*models.Attachment, error),
) *AttachmentUsecaseAdapter {
	return &AttachmentUsecaseAdapter{
		uploadFunc: uploadFunc,
	}
}

func (a *AttachmentUsecaseAdapter) UploadAttachment(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileData []byte, hasPosition bool, position int) (*models.Attachment, error) {
	reader := bytes.NewReader(fileData)
	return a.uploadFunc(ctx, noteID, userID, fileName, fileSize, mimeType, reader, hasPosition, position)
}
