package grpc

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	notesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// NoteGrpcClient клиент для gRPC сервера заметок
type NoteGrpcClient struct {
	client notesGrpc.NoteServiceClient
	conn   *grpc.ClientConn
}

// NewNoteGrpcClient создает новый gRPC клиент
func NewNoteGrpcClient(addr string, opts ...grpc.DialOption) (*NoteGrpcClient, error) {
	if opts == nil {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &NoteGrpcClient{
		client: notesGrpc.NewNoteServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает соединение
func (c *NoteGrpcClient) Close() error {
	return c.conn.Close()
}

// addUserIDToContext добавляет userID в метаданные gRPC
func (c *NoteGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	return metadata.NewOutgoingContext(ctx, md)
}

// ========== NOTE METHODS ==========

// GetNotes получение всех заметок пользователя
func (c *NoteGrpcClient) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetNotes(ctxWithUserID, &notesGrpc.GetNotesRequest{})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoNotes(resp.GetNotes()), nil
}

// GetNote получение заметки по ID
func (c *NoteGrpcClient) GetNote(ctx context.Context, noteID, userID uuid.UUID) (*models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetNote(ctxWithUserID, &notesGrpc.GetNoteRequest{
		NoteId: noteID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoNote(resp.GetNote()), nil
}

// CreateNote создание заметки
func (c *NoteGrpcClient) CreateNote(ctx context.Context, userID uuid.UUID, title string, parentID *uuid.UUID) (*models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	req := &notesGrpc.CreateNoteRequest{
		Title: title,
	}

	if parentID != nil {
		parentIDStr := parentID.String()
		req.ParentId = &parentIDStr
	}

	resp, err := c.client.CreateNote(ctxWithUserID, req)
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoNote(resp), nil
}

// UpdateNote обновление заметки
func (c *NoteGrpcClient) UpdateNote(ctx context.Context, noteID, userID uuid.UUID, title string, parentID *uuid.UUID) (*models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	req := &notesGrpc.UpdateNoteRequest{
		NoteId: noteID.String(),
		Title:  title,
	}

	if parentID != nil {
		parentIDStr := parentID.String()
		req.ParentId = &parentIDStr
	}

	resp, err := c.client.UpdateNote(ctxWithUserID, req)
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoNote(resp), nil
}

// DeleteNote удаление заметки
func (c *NoteGrpcClient) DeleteNote(ctx context.Context, noteID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteNote(ctxWithUserID, &notesGrpc.DeleteNoteRequest{
		NoteId: noteID.String(),
	})
	return c.handleError(err)
}

// ========== BLOCK METHODS ==========

// CreateBlock создание блока
func (c *NoteGrpcClient) CreateBlock(ctx context.Context, noteID, userID uuid.UUID, blockTypeID, position int, content string) (*models.Block, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.CreateBlock(ctxWithUserID, &notesGrpc.CreateBlockRequest{
		NoteId:      noteID.String(),
		BlockTypeId: int32(blockTypeID),
		Position:    int32(position),
		Content:     content,
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlock(resp), nil
}

// GetBlock получение блока
func (c *NoteGrpcClient) GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.Block, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetBlock(ctxWithUserID, &notesGrpc.GetBlockRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlock(resp), nil
}

// UpdateBlockContent обновление контента блока
func (c *NoteGrpcClient) UpdateBlockContent(ctx context.Context, blockID, noteID, userID uuid.UUID, content string) (*models.Block, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.UpdateBlockContent(ctxWithUserID, &notesGrpc.UpdateBlockContentRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
		Content: content,
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlock(resp), nil
}

// MoveBlock перемещение блока
func (c *NoteGrpcClient) MoveBlock(ctx context.Context, blockID, noteID, userID uuid.UUID, newPosition int) (*models.Block, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.MoveBlock(ctxWithUserID, &notesGrpc.MoveBlockRequest{
		NoteId:      noteID.String(),
		BlockId:     blockID.String(),
		NewPosition: int32(newPosition),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlock(resp), nil
}

// DeleteBlock удаление блока
func (c *NoteGrpcClient) DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteBlock(ctxWithUserID, &notesGrpc.DeleteBlockRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	return c.handleError(err)
}

// ========== FORMATTING METHODS ==========

// UpdateBlockFormatting обновление форматирования блока
func (c *NoteGrpcClient) UpdateBlockFormatting(ctx context.Context, blockID, noteID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.UpdateBlockFormatting(ctxWithUserID, &notesGrpc.UpdateFormattingRequest{
		NoteId:     noteID.String(),
		BlockId:    blockID.String(),
		Formatting: ToProtoFormattingRange(&formattingRange),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlockFormatting(resp), nil
}

// ResetBlockFormatting сброс форматирования блока
func (c *NoteGrpcClient) ResetBlockFormatting(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.BlockFormatting, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.ResetBlockFormatting(ctxWithUserID, &notesGrpc.ResetFormattingRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlockFormatting(resp), nil
}

// GetBlockFormatting получение форматирования блока
func (c *NoteGrpcClient) GetBlockFormatting(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.BlockFormatting, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetBlockFormatting(ctxWithUserID, &notesGrpc.GetBlockFormattingRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoBlockFormatting(resp), nil
}

// GetBlocksWithFormatting получение всех блоков с форматированием
func (c *NoteGrpcClient) GetBlocksWithFormatting(ctx context.Context, noteID uuid.UUID) ([]models.Block, map[string]models.BlockFormatting, error) {
	resp, err := c.client.GetBlocksWithFormatting(ctx, &notesGrpc.GetBlocksWithFormattingRequest{
		NoteId: noteID.String(),
	})
	if err != nil {
		return nil, nil, c.handleError(err)
	}

	blocks := make([]models.Block, 0, len(resp.GetBlocks()))
	blockFormattings := make(map[string]models.BlockFormatting)

	for _, protoBlockWF := range resp.GetBlocks() {
		block, formatting := FromProtoBlockWithFormatting(protoBlockWF)
		if block != nil {
			blocks = append(blocks, *block)
			if formatting != nil {
				blockFormattings[block.ID.String()] = *formatting
			}
		}
	}

	return blocks, blockFormattings, nil
}

// ========== SUBNOTE METHODS ==========

// GetSubnotes получение подзаметок
func (c *NoteGrpcClient) GetSubnotes(ctx context.Context, noteID, userID uuid.UUID) ([]models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetSubnotes(ctxWithUserID, &notesGrpc.GetSubnotesRequest{
		NoteId: noteID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoNotes(resp.GetSubnotes()), nil
}

// CreateSubnote создание подзаметки
func (c *NoteGrpcClient) CreateSubnote(ctx context.Context, parentNoteID, userID uuid.UUID, title string) (*models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.CreateSubnote(ctxWithUserID, &notesGrpc.CreateSubnoteRequest{
		ParentNoteId: parentNoteID.String(),
		Title:        title,
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoNote(resp), nil
}

// DeleteSubnote удаление подзаметки
func (c *NoteGrpcClient) DeleteSubnote(ctx context.Context, noteID, subnoteID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteSubnote(ctxWithUserID, &notesGrpc.DeleteSubnoteRequest{
		NoteId:    noteID.String(),
		SubnoteId: subnoteID.String(),
	})
	return c.handleError(err)
}

// handleError обрабатывает gRPC ошибки
func (c *NoteGrpcClient) handleError(err error) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return notes.ErrNoteNotFound
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return notes.ErrNoteNotFound
	case codes.PermissionDenied:
		return notes.ErrForbidden
	case codes.InvalidArgument:
		return notes.ErrInvalidNoteData
	case codes.AlreadyExists:
		return notes.ErrNoteNotFound
	case codes.Internal:
		return notes.ErrNoteNotFound
	default:
		return err
	}
}
