package grpcserver

import (
	"context"

	attachmentsgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) error
}

type Server struct {
	attachmentsgrpc.UnimplementedAttachmentServiceServer
	repo AttachmentRepository
}

func NewServer(repo AttachmentRepository) *Server {
	return &Server{repo: repo}
}

func (s *Server) GetAttachment(ctx context.Context, req *attachmentsgrpc.GetAttachmentRequest) (*attachmentsgrpc.AttachmentResponse, error) {
	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}

	attachment, err := s.repo.GetAttachment(ctx, blockID)
	if err != nil {
		return nil, err
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

	if err := s.repo.DeleteAttachment(ctx, blockID); err != nil {
		return nil, err
	}

	return &attachmentsgrpc.DeleteAttachmentResponse{}, nil
}
