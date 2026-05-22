package grpcclient

import (
	"context"

	notesgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type NotesServiceClient interface {
	GetNote(ctx context.Context, noteID, userID uuid.UUID) (*notesgen.NoteResponse, error)
	GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error)
	GetBlocks(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error)
	CreateBlock(ctx context.Context, userID uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error
	DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (uuid.UUID, error)
	Close() error
}

type notesServiceClient struct {
	client notesgen.NoteServiceClient
	conn   *grpc.ClientConn
}

func NewNotesServiceClient(addr string) (NotesServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &notesServiceClient{
		client: notesgen.NewNoteServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *notesServiceClient) GetNote(ctx context.Context, noteID, userID uuid.UUID) (*notesgen.NoteResponse, error) {
	return c.client.GetNote(ctx, &notesgen.GetNoteRequest{
		NoteId: noteID.String(),
		UserId: userID.String(),
	})
}

func (c *notesServiceClient) GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error) {
	return c.client.GetBlock(ctx, &notesgen.GetBlockRequest{
		BlockId: blockID.String(),
		NoteId:  noteID.String(),
		UserId:  userID.String(),
	})
}

func (c *notesServiceClient) GetBlocks(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
	resp, err := c.client.GetBlocks(ctx, &notesgen.GetBlocksRequest{
		NoteId: noteID.String(),
		UserId: userID.String(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Blocks, nil
}

func (c *notesServiceClient) CreateBlock(ctx context.Context, userID uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
	return c.client.CreateBlock(ctx, &notesgen.CreateBlockRequest{
		UserId: userID.String(),
		Block:  block,
	})
}

func (c *notesServiceClient) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
	_, err := c.client.ShiftBlockPositions(ctx, &notesgen.ShiftBlockPositionsRequest{
		NoteId:       noteID.String(),
		FromPosition: int32(fromPosition),
		Direction:    int32(direction),
	})
	return err
}

func (c *notesServiceClient) DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (uuid.UUID, error) {
	resp, err := c.client.DeleteBlock(ctx, &notesgen.DeleteBlockRequest{
		BlockId: blockID.String(),
		NoteId:  noteID.String(),
		UserId:  userID.String(),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(resp.NoteId)
}

func (c *notesServiceClient) Close() error {
	return c.conn.Close()
}
