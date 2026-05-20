package usecase

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-playground/validator/v10"
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

func (u *authUsecase) SignupUser(ctx context.Context, username, password string) (userID string, err error) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return "", auth.ErrInvalidUsername
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return "", auth.ErrInvalidPassword
	}

	userUUID, err := u.profilesClient.SignupUser(ctx, username, password)
	if err != nil {
		if err == auth.ErrUserExist {
			return "", auth.ErrUserExist
		}
		return "", err
	}

	return userUUID.String(), nil
}

func (u *authUsecase) SigninUser(ctx context.Context, username, password string) (userID string, err error) {
	userUUID, passwordHash, err := u.profilesClient.SigninUser(ctx, username)
	if err != nil {
		if err == auth.ErrUserNotExist {
			return "", auth.ErrUserNotExist
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return "", auth.ErrBadCredentials
	}

	return userUUID.String(), nil
}

func (u *authUsecase) Logout(ctx context.Context, w http.ResponseWriter) {
	auth.DeleteCookie(w, u.jwtConfig.CookieName, u.jwtConfig.Secure)
}
