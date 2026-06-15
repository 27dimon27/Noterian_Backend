package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type OnboardingSeeder interface {
	SeedOnboardingNote(ctx context.Context, userID uuid.UUID) error
}

type authUsecase struct {
	profilesClient grpcclient.ProfilesServiceClient
	jwtConfig      config.JWTConfig
	validate       *validator.Validate
	onboarding     OnboardingSeeder
	logger         *slog.Logger
}

func NewAuthUsecase(profilesClient grpcclient.ProfilesServiceClient, jwtConfig config.JWTConfig, onboarding OnboardingSeeder, logger *slog.Logger) (*authUsecase, error) {
	validate := validator.New()
	if err := initValidator(validate); err != nil {
		return nil, err
	}

	return &authUsecase{
		profilesClient: profilesClient,
		jwtConfig:      jwtConfig,
		validate:       validate,
		onboarding:     onboarding,
		logger:         logger,
	}, nil
}

func (u *authUsecase) SignupUser(ctx context.Context, username, password string) (*dto.Profile, types.AppErrorInterface) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return nil, &types.AppError{
			Err:        auth.ErrInvalidUsername,
			PublicMsg:  auth.PublicMsgErrInvalidUsername,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "SignupUser",
		}
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, &types.AppError{
			Err:        auth.ErrInvalidPassword,
			PublicMsg:  auth.PublicMsgErrInvalidPassword,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "SignupUser",
		}
	}

	profile, err := u.profilesClient.SignupUser(ctx, username, password)
	if err != nil {
		return nil, &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  auth.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "SignupUser",
		}
	}

	userID, err := uuid.Parse(profile.GetId())
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  auth.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "SignupUser",
		}
	}

	if u.onboarding != nil {
		if err := u.onboarding.SeedOnboardingNote(ctx, userID); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  auth.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "usecase",
				Op:         "SignupUser",
			}
		}
	}

	return &dto.Profile{
		ID:       userID,
		Username: profile.GetUsername(),
		Avatar:   profile.GetAvatar(),
	}, nil
}

func (u *authUsecase) SigninUser(ctx context.Context, username, password string) (*dto.Profile, types.AppErrorInterface) {
	profile, err := u.profilesClient.SigninUser(ctx, username)
	if err != nil {
		return nil, &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  auth.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "SigninUser",
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(profile.GetPassword()), []byte(password)); err != nil {
		return nil, &types.AppError{
			Err:        auth.ErrBadCredentials,
			PublicMsg:  auth.PublicMsgErrBadCredentials,
			StatusCode: 401,
			Layer:      "usecase",
			Op:         "SigninUser",
		}
	}

	userID, err := uuid.Parse(profile.GetId())
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  auth.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "SigninUser",
		}
	}

	return &dto.Profile{
		ID:       userID,
		Username: profile.GetUsername(),
		Avatar:   profile.GetAvatar(),
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, w http.ResponseWriter) {
	auth.DeleteCookie(w, u.jwtConfig.CookieName, u.jwtConfig.Secure)
}
