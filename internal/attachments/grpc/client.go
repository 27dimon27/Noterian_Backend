package grpc

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	attachmentsGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AttachmentGrpcClient struct {
	client attachmentsGrpc.AttachmentServiceClient
	conn   *grpc.ClientConn
}

func NewAttachmentGrpcClient(addr string, opts ...grpc.DialOption) (*AttachmentGrpcClient, error) {
	if opts == nil {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &AttachmentGrpcClient{
		client: attachmentsGrpc.NewAttachmentServiceClient(conn),
		conn:   conn,
	}, nil
}

// func (c *AttachmentGrpcClient) Close() error {
// 	return c.conn.Close()
// }

func (c *AttachmentGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	if token, ok := ctx.Value(types.JWTTokenKey).(string); ok && token != "" {
		md = metadata.Join(md, metadata.Pairs("authorization", "Bearer "+token, "token", token))
	}
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *AttachmentGrpcClient) GetAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetAttachment(ctxWithUserID, &attachmentsGrpc.GetAttachmentRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoAttachment(resp), nil
}

func (c *AttachmentGrpcClient) UploadAttachment(
	ctx context.Context,
	noteID, blockID, userID uuid.UUID,
	fileReader io.Reader,
	fileName string,
	fileSize int64,
	mimeType string,
) (*models.Attachment, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	stream, err := c.client.UploadAttachment(ctxWithUserID)
	if err != nil {
		return nil, err
	}

	err = stream.Send(&attachmentsGrpc.UploadAttachmentRequest{
		Data: &attachmentsGrpc.UploadAttachmentRequest_Metadata{
			Metadata: &attachmentsGrpc.FileMetadata{
				NoteId:   noteID.String(),
				BlockId:  blockID.String(),
				FileName: fileName,
				FileSize: fileSize,
				MimeType: mimeType,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 64*1024)
	for {
		n, err := fileReader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		err = stream.Send(&attachmentsGrpc.UploadAttachmentRequest{
			Data: &attachmentsGrpc.UploadAttachmentRequest_Chunk{
				Chunk: buf[:n],
			},
		})
		if err != nil {
			return nil, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoAttachment(resp), nil
}

func (c *AttachmentGrpcClient) DeleteAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteAttachment(ctxWithUserID, &attachmentsGrpc.DeleteAttachmentRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	return c.handleError(err)
}

func (c *AttachmentGrpcClient) handleError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return attachments.ErrAttachmentNotFound
	case codes.PermissionDenied:
		return attachments.ErrForbidden
	case codes.AlreadyExists:
		return attachments.ErrBlockAlreadyHasAttach
	case codes.InvalidArgument:
		return attachments.ErrInvalidMimeType
	case codes.ResourceExhausted:
		return attachments.ErrFileTooLarge
	case codes.Unauthenticated:
		return attachments.ErrInvalidUserID
	default:
		return err
	}
}
