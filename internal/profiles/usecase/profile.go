package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
}

type profileUsecase struct {
	profileRepo ProfileRepository
}

func NewProfileUsecase(profileRepo ProfileRepository) handler.ProfileUsecase {
	return &profileUsecase{
		profileRepo: profileRepo,
	}
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	profile, err := u.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
