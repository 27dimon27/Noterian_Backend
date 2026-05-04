package dto

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type NoteRequest struct {
	UserID   uuid.UUID  `json:"-"`
	Title    string     `json:"title"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	IsPublic bool       `json:"is_public"`
}

func FromNoteRequestDTO(noteReq NoteRequest) models.Note {
	return models.Note{
		UserID:   noteReq.UserID,
		Title:    noteReq.Title,
		ParentID: noteReq.ParentID,
		IsPublic: noteReq.IsPublic,
	}
}
