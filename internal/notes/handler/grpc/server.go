package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NoteUsecase interface {
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error)
	GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error)
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) error
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
}

type Server struct {
	notesgrpc.UnimplementedNoteServiceServer
	noteUsecase NoteUsecase
}

func NewServer(noteUsecase NoteUsecase) *Server {
	return &Server{noteUsecase: noteUsecase}
}

func (s *Server) GetNote(ctx context.Context, req *notesgrpc.GetNoteRequest) (*notesgrpc.NoteResponse, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	note, _, _, err := s.noteUsecase.GetNote(ctx, noteID, userID)
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

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	block, err := s.noteUsecase.GetBlock(ctx, blockID, noteID, userID)
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

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	_, blocks, _, err := s.noteUsecase.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	pbBlocks := make([]*notesgrpc.BlockResponse, 0, len(blocks))
	for _, block := range blocks {
		pbBlocks = append(pbBlocks, &notesgrpc.BlockResponse{
			Id:          block.ID.String(),
			NoteId:      block.NoteID.String(),
			BlockTypeId: int32(block.BlockTypeID),
			Position:    int32(block.Position),
			Content:     block.Content,
			CreatedAt:   timestamppb.New(block.CreatedAt),
			UpdatedAt:   timestamppb.New(block.UpdatedAt),
		})
	}

	return &notesgrpc.GetBlocksResponse{Blocks: pbBlocks}, nil
}

func (s *Server) CreateBlock(ctx context.Context, req *notesgrpc.CreateBlockRequest) (*notesgrpc.BlockResponse, error) {
	if req.GetBlock() == nil {
		return nil, errMissingBlock
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetBlock().GetNoteId())
	if err != nil {
		return nil, err
	}

	block := models.Block{
		NoteID:      noteID,
		BlockTypeID: int(req.GetBlock().GetBlockTypeId()),
		Position:    int(req.GetBlock().GetPosition()),
		Content:     req.GetBlock().GetContent(),
	}

	created, err := s.noteUsecase.CreateBlock(ctx, noteID, userID, block)
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

	if err = s.noteUsecase.ShiftBlockPositions(ctx, noteID, int(req.GetFromPosition()), int(req.GetDirection())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteBlock(ctx context.Context, req *notesgrpc.DeleteBlockRequest) (*notesgrpc.DeleteBlockResponse, error) {
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

	if err := s.noteUsecase.DeleteBlock(ctx, blockID, noteID, userID); err != nil {
		return nil, err
	}

	return &notesgrpc.DeleteBlockResponse{NoteId: noteID.String()}, nil
}

type grpcError string

func (e grpcError) Error() string { return string(e) }

const (
	errMissingBlock = grpcError("block payload is required")
	// errMissingNoteID = grpcError("note id was not returned")
)
