package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	attachmentsgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/attachments/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -source=server.go -destination=mocks/mock_usecase.go -package=mocks

type AttachmentUsecase interface {
	GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, types.AppErrorInterface)
	DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
	GetHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Header, types.AppErrorInterface)
	DeleteHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
}

type Server struct {
	attachmentsgrpc.UnimplementedAttachmentServiceServer
	attachmentUsecase AttachmentUsecase
}

func NewServer(attachmentUsecase AttachmentUsecase) *Server {
	return &Server{attachmentUsecase: attachmentUsecase}
}

func (s *Server) GetAttachment(ctx context.Context, req *attachmentsgrpc.GetAttachmentRequest) (*attachmentsgrpc.AttachmentResponse, error) {
	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	attachment, customErr := s.attachmentUsecase.GetAttachment(ctx, noteID, blockID, userID)
	if customErr != nil {
		return nil, customErr.Unwrap()
	}

	if attachment == nil {
		return nil, status.Error(codes.NotFound, "attachment not found")
	}

	return &attachmentsgrpc.AttachmentResponse{
		Id:           attachment.ID.String(),
		BlockId:      attachment.BlockID.String(),
		MinioKey:     attachment.MinioKey,
		AttachUrl:    attachment.AttachURL,
		UrlExpiresAt: attachment.URLExpiresAt.Unix(),
		CreatedAt:    attachment.CreatedAt.Unix(),
		UpdatedAt:    attachment.UpdatedAt.Unix(),
	}, nil
}

func (s *Server) DeleteAttachment(ctx context.Context, req *attachmentsgrpc.DeleteAttachmentRequest) (*attachmentsgrpc.DeleteAttachmentResponse, error) {
	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	if err := s.attachmentUsecase.DeleteAttachment(ctx, noteID, blockID, userID); err != nil {
		return nil, err.Unwrap()
	}

	return &attachmentsgrpc.DeleteAttachmentResponse{}, nil
}

func (s *Server) GetHeader(ctx context.Context, req *attachmentsgrpc.GetHeaderRequest) (*attachmentsgrpc.HeaderResponse, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	header, customErr := s.attachmentUsecase.GetHeader(ctx, noteID, userID)
	if customErr != nil {
		return nil, customErr.Unwrap()
	}

	if header == nil {
		return nil, status.Error(codes.NotFound, "header not found")
	}

	return &attachmentsgrpc.HeaderResponse{
		Id:           header.ID.String(),
		MinioKey:     header.MinioKey,
		HeaderUrl:    header.HeaderURL,
		UrlExpiresAt: header.URLExpiresAt.Unix(),
		CreatedAt:    header.CreatedAt.Unix(),
		UpdatedAt:    header.UpdatedAt.Unix(),
	}, nil
}

func (s *Server) DeleteHeader(ctx context.Context, req *attachmentsgrpc.DeleteHeaderRequest) (*attachmentsgrpc.DeleteHeaderResponse, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	if err := s.attachmentUsecase.DeleteHeader(ctx, noteID, userID); err != nil {
		return nil, err.Unwrap()
	}

	return &attachmentsgrpc.DeleteHeaderResponse{}, nil
}
