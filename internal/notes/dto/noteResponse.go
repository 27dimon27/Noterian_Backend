package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type NoteResponse struct {
	Note   Note    `json:"note"`
	Blocks []Block `json:"blocks"`
}

func ToNoteResponse(note *models.Note, blocks []models.Block) NoteResponse {
	dtoNote := ToNoteDTO(note)

	dtoBlocks := make([]Block, len(blocks))
	for i, block := range blocks {
		dtoBlocks[i] = ToBlockDTO(block)
	}

	return NoteResponse{
		Note:   dtoNote,
		Blocks: dtoBlocks,
	}
}
