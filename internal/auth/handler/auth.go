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
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_handler_auth.go -package=mocks

type AuthClient interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username, password string) (*models.Profile, error)
	LogoutUser(ctx context.Context, userID uuid.UUID) error
}

type AuthHandler struct {
	authClient AuthClient
	jwtConfig  config.JWTConfig
}

func NewAuthHandler(authClient AuthClient, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
		jwtConfig:  jwtConfig,
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

	signUpUser.Username = strings.TrimSpace(signUpUser.Username)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	user, err := h.authClient.SignupUser(r.Context(), signUpUser.Username, signUpUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserExist):
			write.JSONErrorResponse(w, http.StatusConflict, auth.ErrUserExist)
		case errors.Is(err, auth.ErrInvalidUsername) || errors.Is(err, auth.ErrInvalidPassword):
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

	signInUser.Username = strings.TrimSpace(signInUser.Username)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	user, err := h.authClient.SigninUser(r.Context(), signInUser.Username, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials):
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrBadCredentials)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

func (h *AuthHandler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if ok {
		_ = h.authClient.LogoutUser(r.Context(), userID)
	}

	auth.DeleteCookie(w, h.jwtConfig.CookieName, h.jwtConfig.Secure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) saveUserCookie(w http.ResponseWriter, user *models.Profile) {
	token, err := jwt.GenerateToken(user.ID.String(), h.jwtConfig.CookieTime, h.jwtConfig.Secret)
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
		MaxAge:   int(h.jwtConfig.CookieTime.Seconds()),
		Path:     "/",
		Domain:   "",
	})

	write.JSONResponse(w, http.StatusOK, dto.UserResponse{
		ID:       user.ID.String(),
		Username: user.Username,
	})
}
