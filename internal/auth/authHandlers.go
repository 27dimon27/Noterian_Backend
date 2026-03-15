package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"

	"github.com/go-playground/validator/v10"
)

const (
	minPasswordLength = 4
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInternal         = errors.New("internal server error")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrBadCredentials   = errors.New("Неверный логин или пароль")
	validate            = validator.New()
	isSecure            = os.Getenv("IS_SECURE") == "true"
)

type UserRepository interface {
	CreateUser(login, password string) (*models.Account, error)
	ValidateUser(login, password string) (*models.Account, error)
}

type Handler struct {
	jwtConfig config.JWTConfig
	users     UserRepository
}

type UserResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
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

func init() {
	validate.RegisterValidation("login", validateLogin)
	validate.RegisterValidation("password", validatePassword)
}

func NewHandler(jwtConfig config.JWTConfig, users UserRepository) *Handler {
	return &Handler{
		jwtConfig: jwtConfig,
		users:     users,
	}
}

func getFromBody[T dto.SignInUser | dto.SignUpUser](r *http.Request, u *T) error {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return err
	}
	return validate.Struct(u)
}

func (a *Handler) saveUserCookie(w http.ResponseWriter, user *models.Account) {
	tokenStr, err := jwt.GenerateToken(user.ID.String(), a.jwtConfig.CookieTimeJWT, a.jwtConfig.Secret)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, jwt.ErrTokenCreation)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     a.jwtConfig.CookieName,
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(a.jwtConfig.CookieTimeJWT.Seconds()),
		Path:     "/",
	})

	helpers.JSONResponse(w, http.StatusOK, UserResponse{
		ID:    user.ID.String(),
		Login: user.Username,
	})
}

func (a *Handler) SignupUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		helpers.JSONErrorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var signUpUser dto.SignUpUser

	if err := getFromBody(r, &signUpUser); err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, ErrInvalidInput)
		return
	}

	signUpUser.Login = strings.TrimSpace(signUpUser.Login)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	user, err := a.users.CreateUser(signUpUser.Login, signUpUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserExist):
			helpers.JSONErrorResponse(w, http.StatusConflict, storage.ErrUserExist)
		default:
			helpers.JSONErrorResponse(w, http.StatusInternalServerError, ErrInternal)
		}
		return
	}

	a.saveUserCookie(w, user)
}

func (a *Handler) SigninUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		helpers.JSONErrorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var signInUser dto.SignInUser

	if err := getFromBody(r, &signInUser); err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, ErrInvalidInput)
		return
	}

	signInUser.Login = strings.TrimSpace(signInUser.Login)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	user, err := a.users.ValidateUser(signInUser.Login, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExist):
			helpers.JSONErrorResponse(w, http.StatusUnauthorized, ErrBadCredentials)
		default:
			helpers.JSONErrorResponse(w, http.StatusInternalServerError, ErrInternal)
		}
		return
	}

	a.saveUserCookie(w, user)
}

func (a *Handler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.jwtConfig.CookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	})

	w.WriteHeader(http.StatusNoContent)
}
