package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Block struct {
	ID          uuid.UUID `json:"id"`
	NoteID      uuid.UUID `json:"note_id"`
	BlockTypeID int       `json:"block_type_id"`
	Position    int       `json:"position"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BlockWithFormatting struct {
	Block
	Formatting BlockFormatting `json:"formatting"`
}

func ToBlockDTO(block models.Block) Block {
	return Block{
		ID:          block.ID,
		NoteID:      block.NoteID,
		BlockTypeID: block.BlockTypeID,
		Position:    block.Position,
		Content:     block.Content,
		CreatedAt:   block.CreatedAt,
		UpdatedAt:   block.UpdatedAt,
	}
}

func ToBlockWithFormattingDTO(block models.Block, formatting models.BlockFormatting) BlockWithFormatting {
	return BlockWithFormatting{
		Block:      ToBlockDTO(block),
		Formatting: ToBlockFormattingDTO(formatting),
	}
}
