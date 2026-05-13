package grpcclient

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	attachmentsgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AttachmentRepositoryClient struct {
	client attachmentsgrpc.AttachmentServiceClient
}

func NewAttachmentRepositoryClient(client attachmentsgrpc.AttachmentServiceClient) *AttachmentRepositoryClient {
	return &AttachmentRepositoryClient{client: client}
}

func (c *AttachmentRepositoryClient) GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error) {
	resp, err := c.client.GetAttachment(ctx, &attachmentsgrpc.GetAttachmentRequest{BlockId: blockID.String()})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, attachments.ErrAttachmentNotFound
		}
		return nil, err
	}

	attachmentID, err := uuid.Parse(resp.GetId())
	if err != nil {
		return nil, err
	}

	blockUUID, err := uuid.Parse(resp.GetBlockId())
	if err != nil {
		return nil, err
	}

	return &models.Attachment{
		ID:           attachmentID,
		BlockID:      blockUUID,
		MinioKey:     resp.GetMinioKey(),
		AttachURL:    resp.GetAttachUrl(),
		URLExpiresAt: time.Unix(resp.GetUrlExpiresAt(), 0),
		CreatedAt:    time.Unix(resp.GetCreatedAt(), 0),
		UpdatedAt:    time.Unix(resp.GetUpdatedAt(), 0),
	}, nil
}

func (c *AttachmentRepositoryClient) DeleteAttachment(ctx context.Context, blockID uuid.UUID) error {
	_, err := c.client.DeleteAttachment(ctx, &attachmentsgrpc.DeleteAttachmentRequest{BlockId: blockID.String()})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return attachments.ErrAttachmentNotFound
		}
		return err
	}

	return nil
}
