package grpc

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	attachmentsGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AttachmentGrpcClient клиент для gRPC сервера вложений
type AttachmentGrpcClient struct {
	client attachmentsGrpc.AttachmentServiceClient
	conn   *grpc.ClientConn
}

// NewAttachmentGrpcClient создает новый gRPC клиент
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

// Close закрывает соединение
func (c *AttachmentGrpcClient) Close() error {
	return c.conn.Close()
}

// addUserIDToContext добавляет userID в метаданные gRPC
func (c *AttachmentGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	return metadata.NewOutgoingContext(ctx, md)
}

// GetAttachment получение вложения
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

// UploadAttachment загрузка вложения (streaming)
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

	// Отправляем метаданные
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

	// Отправляем файл чанками по 64KB
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

// DeleteAttachment удаление вложения
func (c *AttachmentGrpcClient) DeleteAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteAttachment(ctxWithUserID, &attachmentsGrpc.DeleteAttachmentRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	return c.handleError(err)
}

// handleError обрабатывает gRPC ошибки
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
