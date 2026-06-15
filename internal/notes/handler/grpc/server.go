package grpc

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NoteUsecase interface {
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, types.AppErrorInterface)
	GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, types.AppErrorInterface)
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, types.AppErrorInterface)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) types.AppErrorInterface
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
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

	note, _, _, customErr := s.noteUsecase.GetNote(ctx, noteID, userID)
	if customErr != nil {
		return nil, customErr.Unwrap()
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
		Icon:      note.Icon,
		HeaderUrl: note.HeaderURL,
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

	block, customErr := s.noteUsecase.GetBlock(ctx, blockID, noteID, userID)
	if customErr != nil {
		return nil, customErr.Unwrap()
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

	_, blocks, _, customErr := s.noteUsecase.GetNote(ctx, noteID, userID)
	if customErr != nil {
		return nil, customErr.Unwrap()
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
		return nil, errors.New("block payload is required")
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

	created, customErr := s.noteUsecase.CreateBlock(ctx, noteID, userID, block)
	if customErr != nil {
		return nil, customErr.Unwrap()
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

	if customErr := s.noteUsecase.ShiftBlockPositions(ctx, noteID, int(req.GetFromPosition()), int(req.GetDirection())); customErr != nil {
		return nil, customErr.Unwrap()
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

	if customErr := s.noteUsecase.DeleteBlock(ctx, blockID, noteID, userID); customErr != nil {
		return nil, customErr
	}

	return &notesgrpc.DeleteBlockResponse{NoteId: noteID.String()}, nil
}
