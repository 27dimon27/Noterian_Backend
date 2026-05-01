package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type NotesResponse struct {
	Notes    []Note            `json:"notes"`
	Subnotes map[string][]Note `json:"subnotes"`
	Total    int               `json:"total"`
}

func ToNotesResponse(notes []models.Note, subnotes map[string][]models.Note) NotesResponse {
	dtoNotes := make([]Note, len(notes))
	dtoSubnotes := make(map[string][]Note)

	for i, note := range notes {
		dtoNotes[i] = ToNoteDTO(&note)
		dtoSubnotes[note.ID.String()] = ToSubnotesDTO(subnotes[note.ID.String()])
	}

	return NotesResponse{
		Notes:    dtoNotes,
		Subnotes: dtoSubnotes,
		Total:    len(dtoNotes),
	}
}
