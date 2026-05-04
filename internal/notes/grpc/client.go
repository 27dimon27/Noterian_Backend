package grpc

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	notesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type NoteGrpcClient struct {
	client notesGrpc.NoteServiceClient
	conn   *grpc.ClientConn
}

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

// func (c *NoteGrpcClient) Close() error {
// 	return c.conn.Close()
// }

func (c *NoteGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	if token, ok := ctx.Value(types.JWTTokenKey).(string); ok && token != "" {
		md = metadata.Join(md, metadata.Pairs("authorization", "Bearer "+token, "token", token))
	}
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *NoteGrpcClient) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, map[string][]models.Note, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetNotes(ctxWithUserID, &notesGrpc.GetNotesRequest{})
	if err != nil {
		return nil, nil, c.handleError(err)
	}

	notes := FromProtoNotes(resp.GetNotes())

	subnotes := make(map[string][]models.Note)
	for parentID, protoSubnotes := range resp.GetSubnotes() {
		subnotes[parentID] = FromProtoNotes(protoSubnotes.GetNotes())
	}

	return notes, subnotes, nil
}

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

func (c *NoteGrpcClient) DeleteNote(ctx context.Context, noteID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteNote(ctxWithUserID, &notesGrpc.DeleteNoteRequest{
		NoteId: noteID.String(),
	})
	return c.handleError(err)
}

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

func (c *NoteGrpcClient) DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteBlock(ctxWithUserID, &notesGrpc.DeleteBlockRequest{
		NoteId:  noteID.String(),
		BlockId: blockID.String(),
	})
	return c.handleError(err)
}

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

func (c *NoteGrpcClient) GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
	resp, err := c.client.GetBlocksWithFormatting(ctx, &notesGrpc.GetBlocksWithFormattingRequest{
		NoteId: noteID.String(),
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	blocks := make([]models.Block, 0, len(resp.GetBlocks()))
	for _, protoBlockWF := range resp.GetBlocks() {
		block, _ := FromProtoBlockWithFormatting(protoBlockWF)
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	return blocks, nil
}

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

func (c *NoteGrpcClient) DeleteSubnote(ctx context.Context, noteID, subnoteID, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteSubnote(ctxWithUserID, &notesGrpc.DeleteSubnoteRequest{
		NoteId:    noteID.String(),
		SubnoteId: subnoteID.String(),
	})
	return c.handleError(err)
}

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
