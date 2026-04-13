package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, error)
}

type profileUsecase struct {
	profileRepo ProfileRepository
}

func NewProfileUsecase(profileRepo ProfileRepository) *profileUsecase {
	return &profileUsecase{
		profileRepo: profileRepo,
	}
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	profile, err := u.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.Password = []byte{}
	return profile, nil
}

func (u *profileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	if profile.Username == "" {
		return nil, profiles.ErrInvalidProfileData
	}

	updatedProfile, err := u.profileRepo.UpdateProfile(ctx, userID, profile)
	if err != nil {
		return nil, err
	}

	return updatedProfile, nil
}

func (u *profileUsecase) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	return u.profileRepo.DeleteProfile(ctx, userID)
}

func (u *profileUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error) {
	user, err := u.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return nil, profiles.ErrWrongPassword
	}

	updatedProfile, err := u.profileRepo.ChangePassword(ctx, userID, newPassword)
	if err != nil {
		return nil, err
	}

	return updatedProfile, err
}
