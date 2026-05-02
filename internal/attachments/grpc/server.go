package grpc

import (
	"bytes"
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	attachmentsGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AttachmentGrpcServer реализует gRPC сервер для вложений
type AttachmentGrpcServer struct {
	attachmentsGrpc.UnimplementedAttachmentServiceServer
	attachmentUsecase AttachmentUsecase
}

// AttachmentUsecase интерфейс бизнес-логики
type AttachmentUsecase interface {
	GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID, fileReader io.Reader, fileName string, fileSize int64, mimeType string) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error
}

// NewAttachmentGrpcServer создает новый gRPC сервер
func NewAttachmentGrpcServer(attachmentUsecase AttachmentUsecase) *AttachmentGrpcServer {
	return &AttachmentGrpcServer{
		attachmentUsecase: attachmentUsecase,
	}
}

// getUserIDFromContext извлекает userID из контекста gRPC
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, attachments.ErrInvalidUserID.Error())
	}
	return userID, nil
}

// GetAttachment получение вложения
func (s *AttachmentGrpcServer) GetAttachment(ctx context.Context, req *attachmentsGrpc.GetAttachmentRequest) (*attachmentsGrpc.Attachment, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, attachments.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, attachments.ErrInvalidBlockID.Error())
	}

	attachment, err := s.attachmentUsecase.GetAttachment(ctx, noteID, blockID, userID)
	if err != nil {
		switch err {
		case attachments.ErrForbidden:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case attachments.ErrNoteNotFound, attachments.ErrBlockNotFound, attachments.ErrAttachmentNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return ToProtoAttachment(attachment), nil
}

// UploadAttachment загрузка вложения (streaming)
func (s *AttachmentGrpcServer) UploadAttachment(stream attachmentsGrpc.AttachmentService_UploadAttachmentServer) error {
	ctx := stream.Context()

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	var metadata *AttachmentFileMetadata
	var buffer bytes.Buffer

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		switch data := req.Data.(type) {
		case *attachmentsGrpc.UploadAttachmentRequest_Metadata:
			metadata = FromProtoAttachmentFileMetadata(data.Metadata)
		case *attachmentsGrpc.UploadAttachmentRequest_Chunk:
			if _, err := buffer.Write(data.Chunk); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}

	if metadata == nil {
		return status.Error(codes.InvalidArgument, "metadata is required")
	}

	attachment, err := s.attachmentUsecase.UploadAttachment(
		ctx,
		metadata.NoteID,
		metadata.BlockID,
		userID,
		&buffer,
		metadata.FileName,
		metadata.FileSize,
		metadata.MimeType,
	)
	if err != nil {
		switch err {
		case attachments.ErrForbidden:
			return status.Error(codes.PermissionDenied, err.Error())
		case attachments.ErrNoteNotFound, attachments.ErrBlockNotFound:
			return status.Error(codes.NotFound, err.Error())
		case attachments.ErrBlockAlreadyHasAttach:
			return status.Error(codes.AlreadyExists, err.Error())
		case attachments.ErrInvalidMimeType:
			return status.Error(codes.InvalidArgument, err.Error())
		case attachments.ErrFileTooLarge:
			return status.Error(codes.ResourceExhausted, err.Error())
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	return stream.SendAndClose(ToProtoAttachment(attachment))
}

// DeleteAttachment удаление вложения
func (s *AttachmentGrpcServer) DeleteAttachment(ctx context.Context, req *attachmentsGrpc.DeleteAttachmentRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, attachments.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, attachments.ErrInvalidBlockID.Error())
	}

	err = s.attachmentUsecase.DeleteAttachment(ctx, noteID, blockID, userID)
	if err != nil {
		switch err {
		case attachments.ErrForbidden:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case attachments.ErrNoteNotFound, attachments.ErrBlockNotFound, attachments.ErrAttachmentNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
