package grpcclient

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/google/uuid"
)

type NoteRepositoryClient struct {
	client notesgrpc.NoteServiceClient
}

func NewNoteRepositoryClient(client notesgrpc.NoteServiceClient) *NoteRepositoryClient {
	return &NoteRepositoryClient{client: client}
}

func (c *NoteRepositoryClient) GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	resp, err := c.client.GetNote(ctx, &notesgrpc.GetNoteRequest{NoteId: noteID.String()})
	if err != nil {
		return nil, err
	}

	return &models.Note{
		ID:        uuid.MustParse(resp.GetId()),
		UserID:    uuid.MustParse(resp.GetUserId()),
		Title:     resp.GetTitle(),
		IsPublic:  resp.GetIsPublic(),
		CreatedAt: resp.GetCreatedAt().AsTime(),
		UpdatedAt: resp.GetUpdatedAt().AsTime(),
	}, nil
}

func (c *NoteRepositoryClient) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	resp, err := c.client.GetBlock(ctx, &notesgrpc.GetBlockRequest{BlockId: blockID.String()})
	if err != nil {
		return nil, err
	}

	return &models.Block{
		ID:          uuid.MustParse(resp.GetId()),
		NoteID:      uuid.MustParse(resp.GetNoteId()),
		BlockTypeID: int(resp.GetBlockTypeId()),
		Position:    int(resp.GetPosition()),
		Content:     resp.GetContent(),
		CreatedAt:   resp.GetCreatedAt().AsTime(),
		UpdatedAt:   resp.GetUpdatedAt().AsTime(),
	}, nil
}

func (c *NoteRepositoryClient) GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
	resp, err := c.client.GetBlocks(ctx, &notesgrpc.GetBlocksRequest{NoteId: noteID.String()})
	if err != nil {
		return nil, err
	}

	blocks := make([]models.Block, 0, len(resp.GetBlocks()))
	for _, item := range resp.GetBlocks() {
		blocks = append(blocks, models.Block{
			ID:          uuid.MustParse(item.GetId()),
			NoteID:      uuid.MustParse(item.GetNoteId()),
			BlockTypeID: int(item.GetBlockTypeId()),
			Position:    int(item.GetPosition()),
			Content:     item.GetContent(),
			CreatedAt:   item.GetCreatedAt().AsTime(),
			UpdatedAt:   item.GetUpdatedAt().AsTime(),
		})
	}

	return blocks, nil
}

func (c *NoteRepositoryClient) CreateBlock(ctx context.Context, block models.Block) (*models.Block, error) {
	resp, err := c.client.CreateBlock(ctx, &notesgrpc.CreateBlockRequest{Block: &notesgrpc.BlockResponse{
		NoteId:      block.NoteID.String(),
		BlockTypeId: int32(block.BlockTypeID),
		Position:    int32(block.Position),
		Content:     block.Content,
	}})
	if err != nil {
		return nil, err
	}

	createdBlock := &models.Block{
		ID:          uuid.MustParse(resp.GetId()),
		NoteID:      uuid.MustParse(resp.GetNoteId()),
		BlockTypeID: int(resp.GetBlockTypeId()),
		Position:    int(resp.GetPosition()),
		Content:     resp.GetContent(),
		CreatedAt:   resp.GetCreatedAt().AsTime(),
		UpdatedAt:   resp.GetUpdatedAt().AsTime(),
	}

	return createdBlock, nil
}

func (c *NoteRepositoryClient) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) error {
	_, err := c.client.ShiftBlockPositions(ctx, &notesgrpc.ShiftBlockPositionsRequest{
		NoteId:       noteID.String(),
		FromPosition: int32(fromPosition),
		Direction:    int32(direction),
	})
	return err
}

func (c *NoteRepositoryClient) DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error) {
	resp, err := c.client.DeleteBlock(ctx, &notesgrpc.DeleteBlockRequest{BlockId: blockID.String()})
	if err != nil {
		return nil, err
	}

	noteID, err := uuid.Parse(resp.GetNoteId())
	if err != nil {
		return nil, err
	}

	return &noteID, nil
}
