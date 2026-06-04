package handler

import (
	"context"
	"errors"
	"log"
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
}

func NewAuthHandler(authUsecase AuthUsecase, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		jwtConfig:   jwtConfig,
	}
}

// SignupUser godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя и устанавливает JWT-cookie сессии.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SignUpUser   true  "Данные регистрации"
// @Success      200      {object}  dto.UserResponse "Пользователь создан, в Set-Cookie возвращается JWT"
// @Failure      400      {object}  map[string]string "Некорректные входные данные"
// @Failure      405      {object}  map[string]string "Пустое тело запроса"
// @Failure      409      {object}  map[string]string "Пользователь уже существует"
// @Failure      500      {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /signup [post]
func (h *AuthHandler) SignupUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in SignupUser: %v", err)
		}
	}()

	var signUpUser dto.SignUpUser

	if err := body.GetBody(r, &signUpUser); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signUpUser.Username = strings.TrimSpace(signUpUser.Username)
	signUpUser.Password = strings.TrimSpace(signUpUser.Password)

	profile, err := h.authUsecase.SignupUser(r.Context(), signUpUser.Username, signUpUser.Password)
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

	h.saveUserCookie(w, profile)
}

// SigninUser godoc
// @Summary      Аутентификация пользователя
// @Description  Проверяет учётные данные и устанавливает JWT-cookie сессии.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SignInUser   true  "Учётные данные"
// @Success      200      {object}  dto.UserResponse "Аутентификация успешна, в Set-Cookie возвращается JWT"
// @Failure      400      {object}  map[string]string "Некорректные входные данные"
// @Failure      401      {object}  map[string]string "Неверный логин или пароль"
// @Failure      405      {object}  map[string]string "Пустое тело запроса"
// @Failure      500      {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /signin [post]
func (h *AuthHandler) SigninUser(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in SigninUser: %v", err)
		}
	}()

	var signInUser dto.SignInUser

	if err := body.GetBody(r, &signInUser); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	signInUser.Username = strings.TrimSpace(signInUser.Username)
	signInUser.Password = strings.TrimSpace(signInUser.Password)

	profile, err := h.authUsecase.SigninUser(r.Context(), signInUser.Username, signInUser.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials), errors.Is(err, auth.ErrUserNotExist):
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrBadCredentials)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, auth.ErrInternal)
		}
		return
	}

	h.saveUserCookie(w, profile)
}

// LogoutUser godoc
// @Summary      Выход пользователя
// @Description  Сбрасывает JWT-cookie текущей сессии.
// @Tags         auth
// @Produce      json
// @Success      204  "Успешный выход, тело ответа отсутствует"
// @Failure      401  {object}  map[string]string  "Неавторизован"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /logout [post]
func (h *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	h.authUsecase.Logout(r.Context(), w)
	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *AuthHandler) saveUserCookie(w http.ResponseWriter, profile *profilesdto.Profile) {
	token, err := jwt.GenerateToken(profile.ID.String(), h.jwtConfig.CookieTime, h.jwtConfig.Secret)
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
	})

	write.JSONResponse(w, http.StatusOK, dto.UserResponse{
		ID:       profile.ID.String(),
		Username: profile.Username,
		Avatar:   profile.Avatar,
	})
}
