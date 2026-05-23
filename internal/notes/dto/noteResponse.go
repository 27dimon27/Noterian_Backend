package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type NoteResponse struct {
	Note   Note                  `json:"note"`
	Blocks []BlockWithFormatting `json:"blocks"`
}

func ToNoteResponse(note *models.Note, blocks []models.Block, blockFormattings map[string]models.BlockFormatting) NoteResponse {
	dtoNote := ToNoteDTO(note)

	dtoBlocks := make([]BlockWithFormatting, len(blocks))
	for i, block := range blocks {
		dtoBlocks[i] = ToBlockWithFormattingDTO(block, blockFormattings[block.ID.String()])
	}

	return NoteResponse{
		Note:   dtoNote,
		Blocks: dtoBlocks,
	}
}
