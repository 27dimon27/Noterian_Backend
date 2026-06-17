package usecase

import (
	"context"
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
	profile, customErr := u.profileRepository.GetProfile(ctx, userID)
	if customErr != nil {
		return nil, customErr
	}

	avatar, customErr := u.profileRepository.GetAvatar(ctx, userID)
	if customErr != nil && !customErr.Is(profiles.ErrAvatarNotFound) {
		return nil, customErr
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}

func (u *profileUsecase) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface) {
	profile, customErr := u.profileRepository.GetProfileByUsername(ctx, username)
	if customErr != nil {
		return nil, customErr
	}

	avatar, customErr := u.profileRepository.GetAvatar(ctx, profile.ID)
	if customErr != nil && !customErr.Is(profiles.ErrAvatarNotFound) {
		return nil, customErr
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

	existingProfile, customErr := u.profileRepository.GetProfileByUsername(ctx, profile.Username)
	if customErr == nil && existingProfile.ID != userID {
		return nil, &types.AppError{
			Err:        profiles.ErrUsernameExists,
			PublicMsg:  profiles.PublicMsgErrUsernameExists,
			StatusCode: 409,
			Layer:      "usecase",
			Op:         "UpdateProfile",
		}
	} else if customErr != nil && !customErr.Is(profiles.ErrUserNotExist) {
		return nil, customErr
	}

	updatedProfile, customErr := u.profileRepository.UpdateProfile(ctx, userID, profile)
	if customErr != nil {
		return nil, customErr
	}

	avatar, customErr := u.profileRepository.GetAvatar(ctx, userID)
	if customErr != nil && !customErr.Is(profiles.ErrAvatarNotFound) {
		return nil, customErr
	}

	if avatar != nil {
		updatedProfile.Avatar = avatar.AvatarURL
	}

	return updatedProfile, nil
}

func (u *profileUsecase) DeleteProfile(ctx context.Context, userID uuid.UUID) types.AppErrorInterface {
	if customErr := u.profileRepository.DeleteAvatar(ctx, userID); customErr != nil && !customErr.Is(profiles.ErrAvatarNotFound) {
		return customErr
	}

	if customErr := u.profileRepository.DeleteProfile(ctx, userID); customErr != nil {
		return customErr
	}

	return nil
}

func (u *profileUsecase) DeleteProfileWithCookie(ctx context.Context, userID uuid.UUID, w http.ResponseWriter, jwtCfg config.JWTConfig) types.AppErrorInterface {
	if customErr := u.DeleteProfile(ctx, userID); customErr != nil {
		return customErr
	}

	profiles.DeleteCookie(w, jwtCfg.CookieName, jwtCfg.Secure)

	return nil
}

func (u *profileUsecase) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, types.AppErrorInterface) {
	avatar, customErr := u.profileRepository.GetAvatar(ctx, profileID)
	if customErr != nil {
		return nil, customErr
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
	avatar, customErr := u.profileRepository.UploadAvatar(ctx, profileID, fileName, fileSize, mimeType, fileReader)
	if customErr != nil {
		return nil, customErr
	}

	return avatar, nil
}

func (u *profileUsecase) DeleteAvatar(ctx context.Context, profileID uuid.UUID) types.AppErrorInterface {
	if customErr := u.profileRepository.DeleteAvatar(ctx, profileID); customErr != nil {
		return customErr
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

	passwordHash, customErr := u.profileRepository.GetPassword(ctx, userID)
	if customErr != nil {
		return nil, customErr
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

	updatedProfile, customErr := u.profileRepository.ChangePassword(ctx, userID, newPassword)
	if customErr != nil {
		return nil, customErr
	}

	avatar, customErr := u.profileRepository.GetAvatar(ctx, userID)
	if customErr != nil && !customErr.Is(profiles.ErrAvatarNotFound) {
		return nil, customErr
	}

	if avatar != nil {
		updatedProfile.Avatar = avatar.AvatarURL
	}

	return updatedProfile, nil
}

func (u *profileUsecase) GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, types.AppErrorInterface) {
	passwordHash, customErr := u.profileRepository.GetPassword(ctx, userID)
	if customErr != nil {
		return nil, customErr
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

	profile, customErr := u.profileRepository.SignupUser(ctx, username, password)
	if customErr != nil {
		return nil, customErr
	}

	return profile, nil
}

func (u *profileUsecase) SigninUser(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface) {
	profile, customErr := u.profileRepository.SigninUser(ctx, username)
	if customErr != nil {
		return nil, customErr
	}

	avatar, customErr := u.profileRepository.GetAvatar(ctx, profile.ID)
	if customErr != nil && !customErr.Is(profiles.ErrAvatarNotFound) {
		return nil, customErr
	}

	if avatar != nil {
		profile.Avatar = avatar.AvatarURL
	}

	return profile, nil
}
