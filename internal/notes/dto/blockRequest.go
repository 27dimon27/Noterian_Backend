package dto

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type BlockRequest struct {
	NoteID      uuid.UUID `json:"note_id"`
	BlockTypeID int       `json:"block_type_id"`
	Position    int       `json:"position"`
	Content     string    `json:"content"`
}

type UpdateBlockContentRequest struct {
	Content string `json:"content"`
}

type MoveBlockRequest struct {
	NewPosition int `json:"new_position"`
}

func FromBlockRequestDTO(blockReq BlockRequest) models.Block {
	return models.Block{
		NoteID:      blockReq.NoteID,
		BlockTypeID: blockReq.BlockTypeID,
		Position:    blockReq.Position,
		Content:     blockReq.Content,
	}
}
