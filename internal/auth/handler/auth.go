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

//go:generate mockgen -source=auth.go -destination=mocks/mock_handler_auth.go -package=mocks

type AuthUsecase interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username, password string) (*models.Profile, error)
	Logout(ctx context.Context, w http.ResponseWriter, jwtCfg config.JWTConfig)
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

// SignupUser godoc
// @Summary Регистрация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignUpUser true "Данные для регистрации"
// @Success 200 {object} dto.UserResponse "Успешная регистрация"
// @Failure 400 {object} map[string]string "Невалидный ввод или невалидные username/password"
// @Failure 401 {object} map[string]string "Неавторизован"
// @Failure 405 {object} map[string]string "Неверный метод"
// @Failure 409 {object} map[string]string "Пользователь уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /signup [post]
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

	user, err := h.authUsecase.SignupUser(r.Context(), signUpUser.Username, signUpUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserExist):
			write.JSONErrorResponse(w, http.StatusConflict, auth.ErrUserExist)
		case errors.Is(err, auth.ErrInvalidUsername), errors.Is(err, auth.ErrInvalidPassword):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

// SigninUser godoc
// @Summary Вход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignInUser true "Данные для входа"
// @Success 200 {object} dto.UserResponse "Успешный вход"
// @Failure 400 {object} map[string]string "Невалидный ввод"
// @Failure 401 {object} map[string]string "Неверный логин или пароль"
// @Failure 405 {object} map[string]string "Неверный метод"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /signin [post]
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

	user, err := h.authUsecase.SigninUser(r.Context(), signInUser.Username, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials), errors.Is(err, auth.ErrUserNotExist):
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrBadCredentials)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, user)
}

// LogOutUser godoc
// @Summary Выход пользователя
// @Tags auth
// @Produce json
// @Success 204 "Успешный выход, тело ответа отсутствует"
// @Failure 401 {object} map[string]string "Неавторизован"
// @Failure 405 {object} map[string]string "Неверный метод"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /logout [post]
func (h *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	h.authUsecase.Logout(r.Context(), w, h.jwtConfig)
	write.JSONResponse(w, http.StatusNoContent, nil)
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
