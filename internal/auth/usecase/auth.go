package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, password string) (*models.Profile, error)
	GetUserByUsername(ctx context.Context, username string) (*models.Profile, error)
}

type authUsecase struct {
	userRepo  UserRepository
	jwtConfig config.JWTConfig
	validate  *validator.Validate
}

func NewAuthUsecase(userRepo UserRepository, jwtConfig config.JWTConfig) (*authUsecase, error) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		return nil, err
	}

	return &authUsecase{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
		validate:  validate,
	}, nil
}

func (u *authUsecase) CreateUser(ctx context.Context, username, password string) (*models.Profile, error) {
	if err := u.validate.Var(username, "required,username"); err != nil {
		return nil, auth.ErrInvalidUsername
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, auth.ErrInvalidPassword
	}

	user, err := u.userRepo.CreateUser(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) ValidateUser(ctx context.Context, username, password string) (*models.Profile, error) {
	user, err := u.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, auth.ErrBadCredentials
	}

	return user, nil
}
