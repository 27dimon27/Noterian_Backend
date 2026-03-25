package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type ProfileResponse struct {
	Profile *models.Profile `json:"profile"`
}
