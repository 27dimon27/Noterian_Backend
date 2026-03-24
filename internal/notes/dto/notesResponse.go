package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type NotesResponse struct {
	Notes []models.Note `json:"notes"`
	Total int           `json:"total"`
}
