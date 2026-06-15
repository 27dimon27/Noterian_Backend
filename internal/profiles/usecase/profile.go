package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=profile.go -destination=mocks/mock_usecase_profile.go -package=mocks

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, types.AppErrorInterface)
	GetProfileByUsername(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, types.AppErrorInterface)
	DeleteProfile(ctx context.Context, userID uuid.UUID) types.AppErrorInterface
	GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, types.AppErrorInterface)
	UploadAvatar(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, types.AppErrorInterface)
	DeleteAvatar(ctx context.Context, profileID uuid.UUID) types.AppErrorInterface
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, types.AppErrorInterface)
	GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, types.AppErrorInterface)
	SignupUser(ctx context.Context, username, password string) (*models.Profile, types.AppErrorInterface)
	SigninUser(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface)
}

type profileUsecase struct {
	profileRepository ProfileRepository
	validate          *validator.Validate
	logger            *slog.Logger
}

func NewProfileUsecase(profileRepository ProfileRepository, logger *slog.Logger) (*profileUsecase, error) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		return nil, err
	}

	return &profileUsecase{
		profileRepository: profileRepository,
		validate:          validate,
		logger:            logger,
	}, nil
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, types.AppErrorInterface) {
	profile, err := u.profileRepository.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, userID)
	if err != nil && !errors.Is(err.Unwrap(), profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}

func (u *profileUsecase) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface) {
	profile, err := u.profileRepository.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, profile.ID)
	if err != nil && !errors.Is(err.Unwrap(), profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}

func (u *profileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, types.AppErrorInterface) {
	if err := u.validate.Var(profile.Username, "required,username"); err != nil {
		return nil, &types.AppError{
			Err:        profiles.ErrInvalidProfileData,
			PublicMsg:  profiles.PublicMsgErrInvalidProfileData,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "UpdateProfile",
		}
	}

	existingProfile, err := u.profileRepository.GetProfileByUsername(ctx, profile.Username)
	if err == nil && existingProfile.ID != userID {
		return nil, &types.AppError{
			Err:        profiles.ErrUsernameExists,
			PublicMsg:  profiles.PublicMsgErrUsernameExists,
			StatusCode: 409,
			Layer:      "usecase",
			Op:         "UpdateProfile",
		}
	} else if err != nil && !errors.Is(err.Unwrap(), profiles.ErrUserNotExist) {
		return nil, err
	}

	updatedProfile, err := u.profileRepository.UpdateProfile(ctx, userID, profile)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, userID)
	if err != nil && !errors.Is(err.Unwrap(), profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		updatedProfile.Avatar = avatar.AvatarURL
	}

	return updatedProfile, nil
}

func (u *profileUsecase) DeleteProfile(ctx context.Context, userID uuid.UUID) types.AppErrorInterface {
	if err := u.profileRepository.DeleteAvatar(ctx, userID); err != nil && !errors.Is(err.Unwrap(), profiles.ErrAvatarNotFound) {
		return err
	}

	if err := u.profileRepository.DeleteProfile(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (u *profileUsecase) DeleteProfileWithCookie(ctx context.Context, userID uuid.UUID, w http.ResponseWriter, jwtCfg config.JWTConfig) types.AppErrorInterface {
	if err := u.DeleteProfile(ctx, userID); err != nil {
		return err
	}

	profiles.DeleteCookie(w, jwtCfg.CookieName, jwtCfg.Secure)

	return nil
}

func (u *profileUsecase) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, types.AppErrorInterface) {
	avatar, err := u.profileRepository.GetAvatar(ctx, profileID)
	if err != nil {
		return nil, err
	}

	if avatar == nil {
		return nil, &types.AppError{
			Err:        profiles.ErrAvatarNotFound,
			PublicMsg:  profiles.PublicMsgErrAvatarNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "GetAvatar",
		}
	}

	return avatar, nil
}

func (u *profileUsecase) UploadAvatar(ctx context.Context,
	profileID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Avatar, types.AppErrorInterface) {
	avatar, err := u.profileRepository.UploadAvatar(ctx, profileID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		return nil, err
	}

	return avatar, nil
}

func (u *profileUsecase) DeleteAvatar(ctx context.Context, profileID uuid.UUID) types.AppErrorInterface {
	if err := u.profileRepository.DeleteAvatar(ctx, profileID); err != nil {
		return err
	}

	return nil
}

func (u *profileUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, types.AppErrorInterface) {
	if err := u.validate.Var(newPassword, "required,password"); err != nil {
		return nil, &types.AppError{
			Err:        profiles.ErrInvalidPasswordData,
			PublicMsg:  profiles.PublicMsgErrInvalidPasswordData,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "ChangePassword",
		}
	}

	passwordHash, err := u.profileRepository.GetPassword(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(oldPassword)); err != nil {
		return nil, &types.AppError{
			Err:        profiles.ErrWrongPassword,
			PublicMsg:  profiles.PublicMsgErrWrongPassword,
			StatusCode: 401,
			Layer:      "usecase",
			Op:         "ChangePassword",
		}
	}

	updatedProfile, err := u.profileRepository.ChangePassword(ctx, userID, newPassword)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, userID)
	if err != nil && !errors.Is(err.Unwrap(), profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		updatedProfile.Avatar = avatar.AvatarURL
	}

	return updatedProfile, nil
}

func (u *profileUsecase) GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, types.AppErrorInterface) {
	passwordHash, err := u.profileRepository.GetPassword(ctx, userID)
	if err != nil {
		return nil, err
	}

	return passwordHash, nil
}

func (u *profileUsecase) SignupUser(ctx context.Context, username, password string) (*models.Profile, types.AppErrorInterface) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return nil, &types.AppError{
			Err:        profiles.ErrInvalidProfileData,
			PublicMsg:  profiles.PublicMsgErrInvalidProfileData,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "SignupUser",
		}
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, &types.AppError{
			Err:        profiles.ErrInvalidProfileData,
			PublicMsg:  profiles.PublicMsgErrInvalidProfileData,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "SignupUser",
		}
	}

	profile, err := u.profileRepository.SignupUser(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *profileUsecase) SigninUser(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface) {
	profile, err := u.profileRepository.SigninUser(ctx, username)
	if err != nil {
		return nil, err
	}

	avatar, err := u.profileRepository.GetAvatar(ctx, profile.ID)
	if err != nil && !errors.Is(err.Unwrap(), profiles.ErrAvatarNotFound) {
		return nil, err
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}
