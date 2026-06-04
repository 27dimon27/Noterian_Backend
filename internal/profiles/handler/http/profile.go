package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/google/uuid"
)

//go:generate mockgen -source=profile.go -destination=mocks/mock_handler_profile.go -package=mocks

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error)
	UploadAvatar(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error)
	DeleteAvatar(ctx context.Context, profileID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error)
}

type ProfileHandler struct {
	profileUsecase ProfileUsecase
	jwtConfig      config.JWTConfig
}

func NewProfileHandler(profileUsecase ProfileUsecase, jwtConfig config.JWTConfig) *ProfileHandler {
	return &ProfileHandler{
		profileUsecase: profileUsecase,
		jwtConfig:      jwtConfig,
	}
}

// GetProfile godoc
// @Summary      Профиль текущего пользователя
// @Tags         profile
// @Produce      json
// @Success      200  {object}  dto.Profile
// @Failure      401  {object}  map[string]string  "Неавторизован"
// @Failure      404  {object}  map[string]string  "Пользователь не найден"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /profile [get]
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	profile, err := h.profileUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, profiles.ErrUserNotExist) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToProfileDTO(*profile)

	write.JSONResponse(w, http.StatusOK, response)
}

// UpdateProfile godoc
// @Summary      Обновить профиль
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        request  body      dto.Profile  true  "Новые данные профиля"
// @Success      200      {object}  dto.Profile
// @Failure      400      {object}  map[string]string  "Некорректные данные"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /profile [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in UpdateProfile: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	var dtoUpdateProfile dto.Profile

	if err := body.GetBody(r, &dtoUpdateProfile); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidProfileData)
		return
	}

	updateProfile := dto.FromProfileDTO(dtoUpdateProfile)

	profile, err := h.profileUsecase.UpdateProfile(r.Context(), userID, updateProfile)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrInvalidProfileData), errors.Is(err, profiles.ErrUsernameExists), errors.Is(err, profiles.ErrUserNotExist):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToProfileDTO(*profile)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteProfile godoc
// @Summary      Удалить аккаунт
// @Description  Удаляет профиль текущего пользователя и сбрасывает сессионную cookie.
// @Tags         profile
// @Produce      json
// @Success      204  "Профиль удалён"
// @Failure      400  {object}  map[string]string  "Пользователь не найден"
// @Failure      401  {object}  map[string]string  "Неавторизован"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /profile [delete]
func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	err := h.profileUsecase.DeleteProfile(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrUserNotExist):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtConfig.CookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   h.jwtConfig.Secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	})

	write.JSONResponse(w, http.StatusNoContent, nil)
}

// GetAvatar godoc
// @Summary      Получить аватар
// @Description  Возвращает аватар текущего пользователя со ссылкой на MinIO.
// @Tags         profile
// @Produce      json
// @Success      200  {object}  dto.Avatar
// @Failure      401  {object}  map[string]string  "Неавторизован"
// @Failure      404  {object}  map[string]string  "Аватар не найден"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /profile/avatar [get]
func (h *ProfileHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	avatar, err := h.profileUsecase.GetAvatar(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrAvatarNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToAvatarDTO(*avatar)

	write.JSONResponse(w, http.StatusOK, response)
}

// UploadAvatar godoc
// @Summary      Загрузить аватар
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Файл изображения"
// @Success      201   {object}  dto.Avatar
// @Failure      400   {object}  map[string]string  "Недопустимый mime-type"
// @Failure      401   {object}  map[string]string  "Неавторизован"
// @Failure      413   {object}  map[string]string  "Файл слишком большой"
// @Failure      500   {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /profile/avatar [post]
func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, profiles.MAX_FILE_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, profiles.ErrFileTooLarge)
		} else {
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file in UploadAvatar: %v", err)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	mimeType := http.DetectContentType(buffer)

	if !profiles.AllowedMimeTypes[mimeType] {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidMimeType)
		return
	}

	avatar, err := h.profileUsecase.UploadAvatar(r.Context(), userID, fileHeader.Filename, fileHeader.Size, mimeType, fileToUpload)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
	}

	response := dto.ToAvatarDTO(*avatar)

	write.JSONResponse(w, http.StatusCreated, response)
}

// DeleteAvatar godoc
// @Summary      Удалить аватар
// @Tags         profile
// @Produce      json
// @Success      204  "Аватар удалён"
// @Failure      401  {object}  map[string]string  "Неавторизован"
// @Failure      404  {object}  map[string]string  "Аватар не найден"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /profile/avatar [delete]
func (h *ProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	if err := h.profileUsecase.DeleteAvatar(r.Context(), userID); err != nil {
		switch {
		case errors.Is(err, profiles.ErrAvatarNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

// ChangePassword godoc
// @Summary      Сменить пароль
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UpdatePassword  true  "Старый и новый пароли"
// @Success      200      {object}  dto.Profile
// @Failure      400      {object}  map[string]string  "Некорректные данные или неверный старый пароль"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      404      {object}  map[string]string  "Пользователь не найден"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /profile/password [put]
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in ChangePassword: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	var dtoUpdatePassword dto.UpdatePassword

	if err := body.GetBody(r, &dtoUpdatePassword); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidPasswordData)
		return
	}

	updatedProfile, err := h.profileUsecase.ChangePassword(r.Context(), userID, dtoUpdatePassword.OldPassword, dtoUpdatePassword.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrUserNotExist):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, profiles.ErrWrongPassword), errors.Is(err, profiles.ErrInvalidPasswordData):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToProfileDTO(*updatedProfile)

	write.JSONResponse(w, http.StatusOK, response)
}
