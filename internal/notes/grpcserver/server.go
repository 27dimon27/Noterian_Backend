package grpcserver

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NoteRepository interface {
	GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	CreateBlock(ctx context.Context, block models.Block) (*models.Block, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) error
	DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error)
}

type Server struct {
	notesgrpc.UnimplementedNoteServiceServer
	repo NoteRepository
}

func NewServer(repo NoteRepository) *Server {
	return &Server{repo: repo}
}

func (s *Server) GetNote(ctx context.Context, req *notesgrpc.GetNoteRequest) (*notesgrpc.NoteResponse, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	note, err := s.repo.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	var parentID *string
	if note.ParentID != nil {
		pid := note.ParentID.String()
		parentID = &pid
	}

	return &notesgrpc.NoteResponse{
		Id:        note.ID.String(),
		UserId:    note.UserID.String(),
		Title:     note.Title,
		ParentId:  parentID,
		IsPublic:  note.IsPublic,
		CreatedAt: timestamppb.New(note.CreatedAt),
		UpdatedAt: timestamppb.New(note.UpdatedAt),
	}, nil
}

func (s *Server) GetBlock(ctx context.Context, req *notesgrpc.GetBlockRequest) (*notesgrpc.BlockResponse, error) {
	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}

	block, err := s.repo.GetBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	return &notesgrpc.BlockResponse{
		Id:          block.ID.String(),
		NoteId:      block.NoteID.String(),
		BlockTypeId: int32(block.BlockTypeID),
		Position:    int32(block.Position),
		Content:     block.Content,
		CreatedAt:   timestamppb.New(block.CreatedAt),
		UpdatedAt:   timestamppb.New(block.UpdatedAt),
	}, nil
}

func (s *Server) GetBlocks(ctx context.Context, req *notesgrpc.GetBlocksRequest) (*notesgrpc.GetBlocksResponse, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	blocks, err := s.repo.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	protoBlocks := make([]*notesgrpc.BlockResponse, 0, len(blocks))
	for _, block := range blocks {
		protoBlocks = append(protoBlocks, &notesgrpc.BlockResponse{
			Id:          block.ID.String(),
			NoteId:      block.NoteID.String(),
			BlockTypeId: int32(block.BlockTypeID),
			Position:    int32(block.Position),
			Content:     block.Content,
			CreatedAt:   timestamppb.New(block.CreatedAt),
			UpdatedAt:   timestamppb.New(block.UpdatedAt),
		})
	}

	return &notesgrpc.GetBlocksResponse{Blocks: protoBlocks}, nil
}

func (s *Server) CreateBlock(ctx context.Context, req *notesgrpc.CreateBlockRequest) (*notesgrpc.BlockResponse, error) {
	if req.GetBlock() == nil {
		return nil, errors.New("block payload is required")
	}

	noteID, err := uuid.Parse(req.GetBlock().GetNoteId())
	if err != nil {
		return nil, err
	}

	created, err := s.repo.CreateBlock(ctx, models.Block{
		NoteID:      noteID,
		BlockTypeID: int(req.GetBlock().GetBlockTypeId()),
		Position:    int(req.GetBlock().GetPosition()),
		Content:     req.GetBlock().GetContent(),
	})
	if err != nil {
		return nil, err
	}

	return &notesgrpc.BlockResponse{
		Id:          created.ID.String(),
		NoteId:      created.NoteID.String(),
		BlockTypeId: int32(created.BlockTypeID),
		Position:    int32(created.Position),
		Content:     created.Content,
		CreatedAt:   timestamppb.New(created.CreatedAt),
		UpdatedAt:   timestamppb.New(created.UpdatedAt),
	}, nil
}

func (s *Server) ShiftBlockPositions(ctx context.Context, req *notesgrpc.ShiftBlockPositionsRequest) (*emptypb.Empty, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	if err := s.repo.ShiftBlockPositions(ctx, noteID, int(req.GetFromPosition()), int(req.GetDirection())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteBlock(ctx context.Context, req *notesgrpc.DeleteBlockRequest) (*notesgrpc.DeleteBlockResponse, error) {
	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}

	noteID, err := s.repo.DeleteBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if noteID == nil {
		return nil, errors.New("note id was not returned")
	}

	return &notesgrpc.DeleteBlockResponse{NoteId: noteID.String()}, nil
}
