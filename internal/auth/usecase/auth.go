package usecase

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_usecase_auth.go -package=mocks

type ProfileRepository interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username string) (*models.Profile, error)
}

type authUsecase struct {
	profileRepository ProfileRepository
	jwtConfig         config.JWTConfig
	validate          *validator.Validate
}

func NewAuthUsecase(profileRepository ProfileRepository, jwtConfig config.JWTConfig) (*authUsecase, error) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		return nil, err
	}

	return &authUsecase{
		profileRepository: profileRepository,
		jwtConfig:         jwtConfig,
		validate:          validate,
	}, nil
}

func (u *authUsecase) SignupUser(ctx context.Context, username, password string) (*models.Profile, error) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return nil, auth.ErrInvalidUsername
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, auth.ErrInvalidPassword
	}

	user, err := u.profileRepository.SignupUser(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) SigninUser(ctx context.Context, username, password string) (*models.Profile, error) {
	user, err := u.profileRepository.SigninUser(ctx, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, auth.ErrBadCredentials
	}

	return user, nil
}

func (u *authUsecase) Logout(ctx context.Context, w http.ResponseWriter, jwtCfg config.JWTConfig) {
	auth.DeleteCookie(w, jwtCfg.CookieName, jwtCfg.Secure)
}
