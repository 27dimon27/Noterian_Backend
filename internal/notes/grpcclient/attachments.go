package grpcclient

import (
	"context"

	attachmentsgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/attachments/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type AttachmentsServiceClient interface {
	GetAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) (*attachmentsgen.AttachmentResponse, error)
	DeleteAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) error
	GetHeader(ctx context.Context, noteID, userID uuid.UUID) (*attachmentsgen.HeaderResponse, error)
	DeleteHeader(ctx context.Context, noteID, userID uuid.UUID) error
	Close() error
}

type attachmentsServiceClient struct {
	client attachmentsgen.AttachmentServiceClient
	conn   *grpc.ClientConn
}

func NewAttachmentsServiceClient(addr string) (AttachmentsServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &attachmentsServiceClient{
		client: attachmentsgen.NewAttachmentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *attachmentsServiceClient) GetAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) (*attachmentsgen.AttachmentResponse, error) {
	return c.client.GetAttachment(ctx, &attachmentsgen.GetAttachmentRequest{
		BlockId: blockID.String(),
		NoteId:  noteID.String(),
		UserId:  userID.String(),
	})
}

func (c *attachmentsServiceClient) DeleteAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) error {
	_, err := c.client.DeleteAttachment(ctx, &attachmentsgen.DeleteAttachmentRequest{
		BlockId: blockID.String(),
		NoteId:  noteID.String(),
		UserId:  userID.String(),
	})
	return err
}

func (c *attachmentsServiceClient) GetHeader(ctx context.Context, noteID, userID uuid.UUID) (*attachmentsgen.HeaderResponse, error) {
	return c.client.GetHeader(ctx, &attachmentsgen.GetHeaderRequest{
		NoteId: noteID.String(),
		UserId: userID.String(),
	})
}

func (c *attachmentsServiceClient) DeleteHeader(ctx context.Context, noteID, userID uuid.UUID) error {
	_, err := c.client.DeleteHeader(ctx, &attachmentsgen.DeleteHeaderRequest{
		NoteId: noteID.String(),
		UserId: userID.String(),
	})
	return err
}

func (c *attachmentsServiceClient) Close() error {
	return c.conn.Close()
}
