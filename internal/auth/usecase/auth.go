package usecase

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	profilesClient grpcclient.ProfilesServiceClient
	jwtConfig      config.JWTConfig
	validate       *validator.Validate
}

func NewAuthUsecase(profilesClient grpcclient.ProfilesServiceClient, jwtConfig config.JWTConfig) (*authUsecase, error) {
	validate := validator.New()
	if err := initValidator(validate); err != nil {
		return nil, err
	}

	return &authUsecase{
		profilesClient: profilesClient,
		jwtConfig:      jwtConfig,
		validate:       validate,
	}, nil
}

func (u *authUsecase) SignupUser(ctx context.Context, username, password string) (*dto.Profile, error) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return nil, auth.ErrInvalidUsername
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, auth.ErrInvalidPassword
	}

	profile, err := u.profilesClient.SignupUser(ctx, username, password)
	if err != nil {
		if err == auth.ErrUserExist {
			return nil, auth.ErrUserExist
		}
		return nil, err
	}

	userID, err := uuid.Parse(profile.GetId())
	if err != nil {
		return nil, err
	}

	return &dto.Profile{
		ID:       userID,
		Username: profile.GetUsername(),
		Avatar:   profile.GetAvatar(),
	}, nil
}

func (u *authUsecase) SigninUser(ctx context.Context, username, password string) (*dto.Profile, error) {
	profile, err := u.profilesClient.SigninUser(ctx, username)
	if err != nil {
		if err == auth.ErrUserNotExist {
			return nil, auth.ErrUserNotExist
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(profile.GetPassword()), []byte(password))
	if err != nil {
		return nil, auth.ErrBadCredentials
	}

	userID, err := uuid.Parse(profile.GetId())
	if err != nil {
		return nil, err
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
