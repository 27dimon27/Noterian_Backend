package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	profilesdto "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_handler_auth.go -package=mocks

type AuthUsecase interface {
	SignupUser(ctx context.Context, username, password string) (*profilesdto.Profile, error)
	SigninUser(ctx context.Context, username, password string) (*profilesdto.Profile, error)
	Logout(ctx context.Context, w http.ResponseWriter)
}

type AuthHandler struct {
	authUsecase AuthUsecase
	jwtConfig   config.JWTConfig
	logger      *slog.Logger
}

func NewAuthHandler(authUsecase AuthUsecase, jwtConfig config.JWTConfig, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		jwtConfig:   jwtConfig,
		logger:      logger,
	}
}

func (h *AuthHandler) SignupUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in SignupUser", "error", err)
		}
	}()

	var signUpUser dto.SignUpUser

	if err := body.GetBody(r, &signUpUser); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signUpUser.Username = strings.TrimSpace(signUpUser.Username)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	profile, err := h.authUsecase.SignupUser(r.Context(), signUpUser.Username, signUpUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserExist):
			h.logger.Warn("User already exists")
			write.JSONErrorResponse(w, http.StatusConflict, auth.ErrUserExist)
		case errors.Is(err, auth.ErrInvalidUsername), errors.Is(err, auth.ErrInvalidPassword):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, profile)
}

func (h *AuthHandler) SigninUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in SigninUser", "error", err)
		}
	}()

	var signInUser dto.SignInUser

	if err := body.GetBody(r, &signInUser); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signInUser.Username = strings.TrimSpace(signInUser.Username)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	profile, err := h.authUsecase.SigninUser(r.Context(), signInUser.Username, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials), errors.Is(err, auth.ErrUserNotExist):
			h.logger.Warn("Wrong credentials")
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrBadCredentials)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, profile)
}

func (h *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	h.authUsecase.Logout(r.Context(), w)
	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *AuthHandler) saveUserCookie(w http.ResponseWriter, profile *profilesdto.Profile) {
	token, err := jwt.GenerateToken(profile.ID.String(), h.jwtConfig.CookieTime, h.jwtConfig.Secret)
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
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
	})

	write.JSONResponse(w, http.StatusOK, dto.UserResponse{
		ID:       profile.ID.String(),
		Username: profile.Username,
		Avatar:   profile.Avatar,
	})
}
