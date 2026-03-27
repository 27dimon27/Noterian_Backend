package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type BlockState struct {
	ID         uuid.UUID              `json:"id"`
	BlockID    uuid.UUID              `json:"block_id"`
	Formatting map[string]interface{} `json:"formatting"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

func ToBlockStateDTO(blockState models.BlockState) BlockState {
	return BlockState{
		ID:         blockState.ID,
		BlockID:    blockState.BlockID,
		Formatting: blockState.Formatting,
		CreatedAt:  blockState.CreatedAt,
		UpdatedAt:  blockState.UpdatedAt,
	}
}
