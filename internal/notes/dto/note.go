package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Note struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Title     string     `json:"title"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func ToNoteDTO(note *models.Note) Note {
	return Note{
		ID:        note.ID,
		UserID:    note.UserID,
		Title:     note.Title,
		ParentID:  note.ParentID,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}
}

func ToSubnotesDTO(subnotes []models.Note) []Note {
	var result []Note
	for _, s := range subnotes {
		subnote := ToNoteDTO(&s)
		result = append(result, subnote)
	}
	return result
}
