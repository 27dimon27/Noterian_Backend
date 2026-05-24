package usecase

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username string) (*models.Profile, error)
}

type profileUsecase struct {
	profileRepository ProfileRepository
	validate          *validator.Validate
}

func NewProfileUsecase(profileRepository ProfileRepository) (*profileUsecase, error) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		return nil, err
	}

	return &profileUsecase{
		profileRepository: profileRepository,
		validate:          validate,
	}, nil
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	profile, err := u.profileRepository.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, userID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}

func (u *profileUsecase) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, error) {
	profile, err := u.profileRepository.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, profile.ID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}

func (u *profileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	if err := u.validate.Var(profile.Username, "required,username"); err != nil {
		return nil, profiles.ErrInvalidProfileData
	}

	existingProfile, err := u.profileRepository.GetProfileByUsername(ctx, profile.Username)
	if err == nil && existingProfile.ID != userID {
		return nil, profiles.ErrUsernameExists
	}
	if !errors.Is(err, profiles.ErrUserNotExist) && err != nil {
		return nil, err
	}

	updatedProfile, err := u.profileRepository.UpdateProfile(ctx, userID, profile)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, userID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		updatedProfile.Avatar = avatar.AvatarURL
	}

	return updatedProfile, nil
}

func (u *profileUsecase) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	err := u.profileRepository.DeleteAvatar(ctx, userID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return err
	}

	err = u.profileRepository.DeleteProfile(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *profileUsecase) DeleteProfileWithCookie(ctx context.Context, userID uuid.UUID, w http.ResponseWriter, jwtCfg config.JWTConfig) error {
	err := u.DeleteProfile(ctx, userID)
	if err != nil {
		return err
	}

	profiles.DeleteCookie(w, jwtCfg.CookieName, jwtCfg.Secure)

	return nil
}

func (u *profileUsecase) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error) {
	avatar, err := u.profileRepository.GetAvatar(ctx, profileID)
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
	avatar, err := u.profileRepository.UploadAvatar(ctx, profileID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		return nil, err
	}
	return avatar, nil
}

func (u *profileUsecase) DeleteAvatar(ctx context.Context, profileID uuid.UUID) error {
	if err := u.profileRepository.DeleteAvatar(ctx, profileID); err != nil {
		return err
	}
	return nil
}

func (u *profileUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error) {
	if err := u.validate.Var(newPassword, "required,password"); err != nil {
		return nil, profiles.ErrInvalidPasswordData
	}

	passwordHash, err := u.profileRepository.GetPassword(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(oldPassword))
	if err != nil {
		return nil, profiles.ErrWrongPassword
	}

	updatedProfile, err := u.profileRepository.ChangePassword(ctx, userID, newPassword)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, userID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		updatedProfile.Avatar = avatar.AvatarURL
	}

	return updatedProfile, nil
}

func (u *profileUsecase) GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	passwordHash, err := u.profileRepository.GetPassword(ctx, userID)
	if err != nil {
		return nil, err
	}
	return passwordHash, nil
}

func (u *profileUsecase) SignupUser(ctx context.Context, username, password string) (*models.Profile, error) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return nil, profiles.ErrInvalidProfileData
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, profiles.ErrInvalidProfileData
	}

	return u.profileRepository.SignupUser(ctx, username, password)
}

func (u *profileUsecase) SigninUser(ctx context.Context, username string) (*models.Profile, error) {
	profile, err := u.profileRepository.SigninUser(ctx, username)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, profile.ID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, err
	}

	profile.Avatar = avatar.AvatarURL

	return profile, nil
}
