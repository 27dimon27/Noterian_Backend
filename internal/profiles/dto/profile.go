package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToProfileDTO(profile *models.Profile) Profile {
	return Profile{
		ID:        profile.ID,
		Username:  profile.Username,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}
}

func FromProfileDTO(profile Profile) *models.Profile {
	return &models.Profile{
		ID:       profile.ID,
		Username: profile.Username,
	}
}
