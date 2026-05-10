package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToProfileDTO(profile models.Profile) Profile {
	return Profile{
		ID:       profile.ID,
		Username: profile.Username,
		Avatar:   profile.Avatar,
	}
}

func FromProfileDTO(profile Profile) models.Profile {
	return models.Profile{
		ID:       profile.ID,
		Username: profile.Username,
	}
}
