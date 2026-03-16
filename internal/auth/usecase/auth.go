package usecase

import (
	"regexp"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/repository"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/go-playground/validator/v10"
)

const (
	minPasswordLength = 4
)

type AuthUsecase interface {
	CreateUser(login, password string) (*models.Account, error)
	ValidateUser(login, password string) (*models.Account, error)
	GenerateToken(userID string) (string, error)
}

type authUsecase struct {
	userRepo  repository.UserRepository
	jwtConfig config.JWTConfig
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtConfig config.JWTConfig) AuthUsecase {
	initValidator()
	return &authUsecase{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func initValidator() {
	auth.Validate.RegisterValidation("login", validateLogin)
	auth.Validate.RegisterValidation("password", validatePassword)
}

func validateLogin(fl validator.FieldLevel) bool {
	login := fl.Field().String()

	validLoginRegex := regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9_.]+$`)
	if !validLoginRegex.MatchString(login) {
		return false
	}

	if strings.HasPrefix(login, "_") || strings.HasPrefix(login, ".") ||
		strings.HasSuffix(login, "_") || strings.HasSuffix(login, ".") {
		return false
	}

	if strings.Contains(login, "__") || strings.Contains(login, "..") ||
		strings.Contains(login, "_.") || strings.Contains(login, "._") {
		return false
	}

	return true
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < minPasswordLength {
		return false
	}

	hasUppercase := regexp.MustCompile(`[A-ZА-Я]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUppercase || !hasDigit {
		return false
	}
	return true
}

func (u *authUsecase) CreateUser(login, password string) (*models.Account, error) {
	if err := auth.Validate.Var(login, "required,login"); err != nil {
		return nil, auth.ErrInvalidLogin
	}

	if err := auth.Validate.Var(password, "required,password"); err != nil {
		return nil, auth.ErrInvalidPassword
	}

	user, err := u.userRepo.CreateUser(login, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) ValidateUser(login, password string) (*models.Account, error) {
	user, err := u.userRepo.ValidateUser(login, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) GenerateToken(userID string) (string, error) {
	tokenStr, err := jwt.GenerateToken(userID, u.jwtConfig.CookieTimeJWT, u.jwtConfig.Secret)
	if err != nil {
		return "", auth.ErrTokenCreation
	}
	return tokenStr, nil
}
