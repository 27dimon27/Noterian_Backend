package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	profilesdto "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_handler_auth.go -package=mocks

type AuthUsecase interface {
	SignupUser(ctx context.Context, username, password string) (*profilesdto.Profile, types.AppErrorInterface)
	SigninUser(ctx context.Context, username, password string) (*profilesdto.Profile, types.AppErrorInterface)
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
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "SignupUser", auth.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "SignupUser", err))
		}
	}()

	var signUpUser dto.SignUpUser

	if err := body.GetBody(r, &signUpUser); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "SignupUser", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, auth.PublicMsgErrInternalServer)
		return
	}

	signUpUser.Username = strings.TrimSpace(signUpUser.Username)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	profile, customErr := h.authUsecase.SignupUser(r.Context(), signUpUser.Username, signUpUser.Password)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), auth.ErrUserExist):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusConflict, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), auth.ErrInvalidUsername), errors.Is(customErr.Unwrap(), auth.ErrInvalidPassword):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	h.saveUserCookie(w, profile)
}

func (h *AuthHandler) SigninUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "SigninUser", auth.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "SigninUser", err))
		}
	}()

	var signInUser dto.SignInUser

	if err := body.GetBody(r, &signInUser); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "SigninUser", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, auth.PublicMsgErrInternalServer)
		return
	}

	signInUser.Username = strings.TrimSpace(signInUser.Username)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	profile, customErr := h.authUsecase.SigninUser(r.Context(), signInUser.Username, signInUser.Password)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), auth.ErrBadCredentials), errors.Is(customErr.Unwrap(), auth.ErrUserNotExist):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusUnauthorized, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
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
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "saveUserCookie", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, auth.PublicMsgErrInternalServer)
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
