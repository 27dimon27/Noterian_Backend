package grpc

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	notesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NoteGrpcServer реализует gRPC сервер для заметок
type NoteGrpcServer struct {
	notesGrpc.UnimplementedNoteServiceServer
	noteUsecase NoteUsecase
}

// NoteUsecase интерфейс бизнес-логики
type NoteUsecase interface {
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note, userID uuid.UUID) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error)
	MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
	ResetBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.BlockFormatting, error)
	GetBlocksWithFormatting(ctx context.Context, noteID uuid.UUID) ([]models.Block, map[string]models.BlockFormatting, error)
	GetBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.BlockFormatting, error)
	GetSubnotes(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) ([]models.Note, error)
	CreateSubnote(ctx context.Context, parentNoteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteSubnote(ctx context.Context, noteID uuid.UUID, subnoteID uuid.UUID, userID uuid.UUID) error
}

// NewNoteGrpcServer создает новый gRPC сервер
func NewNoteGrpcServer(noteUsecase NoteUsecase) *NoteGrpcServer {
	return &NoteGrpcServer{
		noteUsecase: noteUsecase,
	}
}

// getUserIDFromContext извлекает userID из контекста gRPC
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, notes.ErrInvalidUserID.Error())
	}
	return userID, nil
}

// ========== NOTE METHODS ==========

// GetNotes получение всех заметок пользователя
func (s *NoteGrpcServer) GetNotes(ctx context.Context, req *notesGrpc.GetNotesRequest) (*notesGrpc.NotesResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	notesList, err := s.noteUsecase.GetNotes(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoNotes := make([]*notesGrpc.Note, 0, len(notesList))
	for i := range notesList {
		protoNotes = append(protoNotes, ToProtoNote(&notesList[i]))
	}

	return &notesGrpc.NotesResponse{
		Notes: protoNotes,
		Total: int32(len(protoNotes)),
	}, nil
}

// GetNote получение заметки по ID
func (s *NoteGrpcServer) GetNote(ctx context.Context, req *notesGrpc.GetNoteRequest) (*notesGrpc.NoteResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	note, err := s.noteUsecase.GetNote(ctx, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	blocks, blockFormattings, err := s.noteUsecase.GetBlocksWithFormatting(ctx, noteID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blocksWithFormatting := make([]*notesGrpc.BlockWithFormatting, 0, len(blocks))
	for _, block := range blocks {
		formatting := blockFormattings[block.ID.String()]
		blocksWithFormatting = append(blocksWithFormatting, ToProtoBlockWithFormatting(&block, &formatting))
	}

	return &notesGrpc.NoteResponse{
		Note:   ToProtoNote(note),
		Blocks: blocksWithFormatting,
	}, nil
}

// CreateNote создание заметки
func (s *NoteGrpcServer) CreateNote(ctx context.Context, req *notesGrpc.CreateNoteRequest) (*notesGrpc.Note, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	note := models.Note{
		UserID: userID,
		Title:  req.GetTitle(),
	}

	if req.ParentId != nil && *req.ParentId != "" {
		parentID, err := uuid.Parse(*req.ParentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
		}
		note.ParentID = &parentID
	}

	createdNote, err := s.noteUsecase.CreateNote(ctx, note)
	if err != nil {
		if errors.Is(err, notes.ErrInvalidNoteData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoNote(createdNote), nil
}

// UpdateNote обновление заметки
func (s *NoteGrpcServer) UpdateNote(ctx context.Context, req *notesGrpc.UpdateNoteRequest) (*notesGrpc.Note, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	note := models.Note{
		Title: req.GetTitle(),
	}

	if req.ParentId != nil && *req.ParentId != "" {
		parentID, err := uuid.Parse(*req.ParentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
		}
		note.ParentID = &parentID
	}

	updatedNote, err := s.noteUsecase.UpdateNote(ctx, noteID, note, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrInvalidNoteData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoNote(updatedNote), nil
}

// DeleteNote удаление заметки
func (s *NoteGrpcServer) DeleteNote(ctx context.Context, req *notesGrpc.DeleteNoteRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	err = s.noteUsecase.DeleteNote(ctx, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// ========== BLOCK METHODS ==========

// CreateBlock создание блока
func (s *NoteGrpcServer) CreateBlock(ctx context.Context, req *notesGrpc.CreateBlockRequest) (*notesGrpc.Block, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	block := models.Block{
		BlockTypeID: int(req.GetBlockTypeId()),
		Position:    int(req.GetPosition()),
		Content:     req.GetContent(),
	}

	createdBlock, err := s.noteUsecase.CreateBlock(ctx, noteID, userID, block)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, notes.ErrInvalidBlockType) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlock(createdBlock), nil
}

// GetBlock получение блока
func (s *NoteGrpcServer) GetBlock(ctx context.Context, req *notesGrpc.GetBlockRequest) (*notesGrpc.Block, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	block, err := s.noteUsecase.GetBlock(ctx, blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlock(block), nil
}

// UpdateBlockContent обновление контента блока
func (s *NoteGrpcServer) UpdateBlockContent(ctx context.Context, req *notesGrpc.UpdateBlockContentRequest) (*notesGrpc.Block, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	updatedBlock, err := s.noteUsecase.UpdateBlockContent(ctx, blockID, noteID, userID, req.GetContent())
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlock(updatedBlock), nil
}

// MoveBlock перемещение блока
func (s *NoteGrpcServer) MoveBlock(ctx context.Context, req *notesGrpc.MoveBlockRequest) (*notesGrpc.Block, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	movedBlock, err := s.noteUsecase.MoveBlock(ctx, blockID, noteID, userID, int(req.GetNewPosition()))
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, notes.ErrInvalidPosition) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlock(movedBlock), nil
}

// DeleteBlock удаление блока
func (s *NoteGrpcServer) DeleteBlock(ctx context.Context, req *notesGrpc.DeleteBlockRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	err = s.noteUsecase.DeleteBlock(ctx, blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// ========== FORMATTING METHODS ==========

// UpdateBlockFormatting обновление форматирования блока
func (s *NoteGrpcServer) UpdateBlockFormatting(ctx context.Context, req *notesGrpc.UpdateFormattingRequest) (*notesGrpc.BlockFormatting, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	formattingRange := FromProtoFormattingRange(req.GetFormatting())
	if formattingRange == nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidFormatting.Error())
	}

	updatedFormatting, err := s.noteUsecase.UpdateBlockFormatting(ctx, blockID, noteID, userID, *formattingRange)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, notes.ErrInvalidBlockType) || errors.Is(err, notes.ErrInvalidFormattingRange) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlockFormatting(updatedFormatting), nil
}

// ResetBlockFormatting сброс форматирования блока
func (s *NoteGrpcServer) ResetBlockFormatting(ctx context.Context, req *notesGrpc.ResetFormattingRequest) (*notesGrpc.BlockFormatting, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	resetFormatting, err := s.noteUsecase.ResetBlockFormatting(ctx, blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlockFormatting(resetFormatting), nil
}

// GetBlockFormatting получение форматирования блока
func (s *NoteGrpcServer) GetBlockFormatting(ctx context.Context, req *notesGrpc.GetBlockFormattingRequest) (*notesGrpc.BlockFormatting, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidBlockID.Error())
	}

	formatting, err := s.noteUsecase.GetBlockFormatting(ctx, blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoBlockFormatting(formatting), nil
}

// GetBlocksWithFormatting получение всех блоков с форматированием
func (s *NoteGrpcServer) GetBlocksWithFormatting(ctx context.Context, req *notesGrpc.GetBlocksWithFormattingRequest) (*notesGrpc.BlocksWithFormattingResponse, error) {
	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	blocks, blockFormattings, err := s.noteUsecase.GetBlocksWithFormatting(ctx, noteID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blocksWithFormatting := make([]*notesGrpc.BlockWithFormatting, 0, len(blocks))
	for _, block := range blocks {
		formatting := blockFormattings[block.ID.String()]
		blocksWithFormatting = append(blocksWithFormatting, ToProtoBlockWithFormatting(&block, &formatting))
	}

	return &notesGrpc.BlocksWithFormattingResponse{
		Blocks: blocksWithFormatting,
	}, nil
}

// ========== SUBNOTE METHODS ==========

// GetSubnotes получение подзаметок
func (s *NoteGrpcServer) GetSubnotes(ctx context.Context, req *notesGrpc.GetSubnotesRequest) (*notesGrpc.SubnotesResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	subnotes, err := s.noteUsecase.GetSubnotes(ctx, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoSubnotes := make([]*notesGrpc.Note, 0, len(subnotes))
	for i := range subnotes {
		protoSubnotes = append(protoSubnotes, ToProtoNote(&subnotes[i]))
	}

	return &notesGrpc.SubnotesResponse{
		Subnotes: protoSubnotes,
	}, nil
}

// CreateSubnote создание подзаметки
func (s *NoteGrpcServer) CreateSubnote(ctx context.Context, req *notesGrpc.CreateSubnoteRequest) (*notesGrpc.Note, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	parentNoteID, err := uuid.Parse(req.GetParentNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	note := models.Note{
		UserID: userID,
		Title:  req.GetTitle(),
	}

	createdNote, err := s.noteUsecase.CreateSubnote(ctx, parentNoteID, userID, note)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoNote(createdNote), nil
}

// DeleteSubnote удаление подзаметки
func (s *NoteGrpcServer) DeleteSubnote(ctx context.Context, req *notesGrpc.DeleteSubnoteRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(req.GetNoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	subnoteID, err := uuid.Parse(req.GetSubnoteId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, notes.ErrInvalidNoteID.Error())
	}

	err = s.noteUsecase.DeleteSubnote(ctx, noteID, subnoteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, notes.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
