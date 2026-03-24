package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type NoteResponse struct {
	Note   *models.Note   `json:"note"`
	Blocks []models.Block `json:"blocks"`
}
