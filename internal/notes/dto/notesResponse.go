package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type NotesResponse struct {
	Notes []Note `json:"notes"`
	Total int    `json:"total"`
}

func ToNotesResponse(notes []models.Note) NotesResponse {
	dtoNotes := make([]Note, len(notes))
	for i, note := range notes {
		dtoNotes[i] = ToNoteDTO(&note)
	}

	return NotesResponse{
		Notes: dtoNotes,
		Total: len(dtoNotes),
	}
}
