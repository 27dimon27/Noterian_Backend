package usecase

import (
	"context"
	"errors"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=profile.go -destination=mocks/mock_usecase_profile.go -package=mocks

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error)
	UploadAvatar(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error)
	DeleteAvatar(ctx context.Context, profileID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, error)
	GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, error)
}

type profileUsecase struct {
	profileRepo ProfileRepository
	validate    *validator.Validate
}

func NewProfileUsecase(profileRepo ProfileRepository) (*profileUsecase, error) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		return nil, err
	}

	return &profileUsecase{
		profileRepo: profileRepo,
		validate:    validate,
	}, nil
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	profile, err := u.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *profileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	if err := u.validate.Var(profile.Username, "required,username"); err != nil {
		return nil, profiles.ErrInvalidProfileData
	}

	_, err := u.profileRepo.GetProfileByUsername(ctx, profile.Username)
	if err == nil {
		return nil, profiles.ErrUsernameExists
	}
	if !errors.Is(err, profiles.ErrUserNotExist) {
		return nil, err
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

func (u *profileUsecase) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error) {
	avatar, err := u.profileRepo.GetAvatar(ctx, profileID)
	if err != nil {
		return nil, err
	}

	if avatar == nil {
		return nil, profiles.ErrAvatarNotFound
	}

	return avatar, nil
}

func (u *profileUsecase) UploadAvatar(ctx context.Context,
	profileID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Avatar, error) {
	avatar, err := u.profileRepo.UploadAvatar(ctx, profileID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		return nil, err
	}
	return avatar, nil
}

func (u *profileUsecase) DeleteAvatar(ctx context.Context, profileID uuid.UUID) error {
	if err := u.profileRepo.DeleteAvatar(ctx, profileID); err != nil {
		return err
	}
	return nil
}

func (u *profileUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error) {
	if err := u.validate.Var(newPassword, "required,password"); err != nil {
		return nil, profiles.ErrInvalidPasswordData
	}

	password, err := u.profileRepo.GetPassword(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(password, []byte(oldPassword))
	if err != nil {
		return nil, profiles.ErrWrongPassword
	}

	updatedProfile, err := u.profileRepo.ChangePassword(ctx, userID, newPassword)
	if err != nil {
		return nil, err
	}

	return updatedProfile, err
}
