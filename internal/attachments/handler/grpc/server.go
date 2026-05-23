package grpc

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	attachmentsgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/attachments/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AttachmentUsecase interface {
	GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error
	GetHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Header, error)
	DeleteHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
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

	attachment, err := s.attachmentUsecase.GetAttachment(ctx, noteID, blockID, userID)
	if err != nil {
		return nil, err
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
		return nil, err
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

	header, err := s.attachmentUsecase.GetHeader(ctx, noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrHeaderNotFound):
			return nil, status.Error(codes.NotFound, "header not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
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
		return nil, err
	}

	return &attachmentsgrpc.DeleteHeaderResponse{}, nil
}
