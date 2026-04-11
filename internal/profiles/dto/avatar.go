package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Avatar struct {
	ID           uuid.UUID
	ProfileID    uuid.UUID
	MinioKey     string
	AvatarURL    string
	URLExpiresAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ToAvatarDTO(avatar models.Avatar) Avatar {
	return Avatar{
		ID:           avatar.ID,
		ProfileID:    avatar.ProfileID,
		MinioKey:     avatar.MinioKey,
		AvatarURL:    avatar.AvatarURL,
		URLExpiresAt: avatar.URLExpiresAt,
		CreatedAt:    avatar.CreatedAt,
		UpdatedAt:    avatar.UpdatedAt,
	}
}
