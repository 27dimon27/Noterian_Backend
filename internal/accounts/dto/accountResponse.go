package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type AccountResponse struct {
	Account *models.Account `json:"account"`
}
