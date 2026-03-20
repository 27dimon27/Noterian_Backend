package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
	jwtConfig   config.JWTConfig
}

type UserResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

func NewAuthHandler(authUsecase usecase.AuthUsecase, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		jwtConfig:   jwtConfig,
	}
}

func getFromBody[T dto.SignInUser | dto.SignUpUser](r *http.Request, u *T) error {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return err
	}
	return nil
}

func (h *AuthHandler) saveUserCookie(w http.ResponseWriter, user *models.Account) {
	token, err := jwt.GenerateToken(user.ID.String(), h.jwtConfig.CookieTimeJWT, h.jwtConfig.Secret)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrTokenCreation)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtConfig.CookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   h.jwtConfig.Secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.jwtConfig.CookieTimeJWT.Seconds()),
		Path:     "/",
	})

	helpers.JSONResponse(w, http.StatusOK, UserResponse{
		ID:    user.ID.String(),
		Login: user.Username,
	})
}

func (h *AuthHandler) SignupUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		helpers.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var signUpUser dto.SignUpUser

	if err := getFromBody(r, &signUpUser); err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signUpUser.Login = strings.TrimSpace(signUpUser.Login)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	user, err := h.authUsecase.CreateUser(signUpUser.Login, signUpUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserExist):
			helpers.JSONErrorResponse(w, http.StatusConflict, auth.ErrUserExist)
		case errors.Is(err, auth.ErrInvalidLogin) || errors.Is(err, auth.ErrInvalidPassword):
			helpers.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			helpers.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

func (h *AuthHandler) SigninUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		helpers.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var signInUser dto.SignInUser

	if err := getFromBody(r, &signInUser); err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signInUser.Login = strings.TrimSpace(signInUser.Login)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	user, err := h.authUsecase.ValidateUser(signInUser.Login, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials) || errors.Is(err, auth.ErrUserNotExist):
			helpers.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrBadCredentials)
		default:
			helpers.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

func (h *AuthHandler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtConfig.CookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   h.jwtConfig.Secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	})

	w.WriteHeader(http.StatusNoContent)
}
