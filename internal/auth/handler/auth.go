package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
)

type AuthUsecase interface {
	CreateUser(ctx context.Context, login, password string) (*models.Profile, error)
	ValidateUser(ctx context.Context, login, password string) (*models.Profile, error)
}

type AuthHandler struct {
	authUsecase AuthUsecase
	jwtConfig   config.JWTConfig
}

func NewAuthHandler(authUsecase AuthUsecase, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		jwtConfig:   jwtConfig,
	}
}

func (h *AuthHandler) SignupUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var signUpUser dto.SignUpUser

	if err := body.GetBody(r, &signUpUser); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signUpUser.Login = strings.TrimSpace(signUpUser.Login)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	user, err := h.authUsecase.CreateUser(r.Context(), signUpUser.Login, signUpUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserExist):
			write.JSONErrorResponse(w, http.StatusConflict, auth.ErrUserExist)
		case errors.Is(err, auth.ErrInvalidLogin) || errors.Is(err, auth.ErrInvalidPassword):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

func (h *AuthHandler) SigninUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var signInUser dto.SignInUser

	if err := body.GetBody(r, &signInUser); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signInUser.Login = strings.TrimSpace(signInUser.Login)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	user, err := h.authUsecase.ValidateUser(r.Context(), signInUser.Login, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials) || errors.Is(err, auth.ErrUserNotExist):
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrBadCredentials)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

func (h *AuthHandler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	auth.DeleteCookie(w, h.jwtConfig.CookieName, h.jwtConfig.Secure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) saveUserCookie(w http.ResponseWriter, user *models.Profile) {
	token, err := jwt.GenerateToken(user.ID.String(), h.jwtConfig.CookieTimeJWT, h.jwtConfig.Secret)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrTokenCreation)
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

	write.JSONResponse(w, http.StatusOK, dto.UserResponse{
		ID:    user.ID.String(),
		Login: user.Username,
	})
}
