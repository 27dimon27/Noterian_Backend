package usecase

import (
	"regexp"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 4
)

type UserRepository interface {
	CreateUser(login, password string) (*models.Account, error)
	GetUserByLogin(login string) (*models.Account, error)
}

type authUsecase struct {
	userRepo  UserRepository
	jwtConfig config.JWTConfig
	validate  *validator.Validate
}

func NewAuthUsecase(userRepo UserRepository, jwtConfig config.JWTConfig) handler.AuthUsecase {
	validate := validator.New()
	initValidator(validate)

	return &authUsecase{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
		validate:  validate,
	}
}

func initValidator(validate *validator.Validate) {
	validate.RegisterValidation("login", validateLogin)
	validate.RegisterValidation("password", validatePassword)
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
	if err := u.validate.Var(login, "required,login"); err != nil {
		return nil, auth.ErrInvalidLogin
	}

	if err := u.validate.Var(password, "required,password"); err != nil {
		return nil, auth.ErrInvalidPassword
	}

	user, err := u.userRepo.CreateUser(login, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) ValidateUser(login, password string) (*models.Account, error) {
	user, err := u.userRepo.GetUserByLogin(login)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, auth.ErrBadCredentials
	}

	return user, nil
}
