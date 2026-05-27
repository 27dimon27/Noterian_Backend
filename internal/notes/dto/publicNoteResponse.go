package dto

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type PublicNoteResponse struct {
    ID    uuid.UUID `json:"id"`
    Title string    `json:"title"`
    Icon  string    `json:"icon"`
}

func ToPublicNoteResponse(note models.Note) PublicNoteResponse {
    return PublicNoteResponse{
        ID:    note.ID,
        Title: note.Title,
        Icon:  note.Icon,
    }
}
