package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
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
// @Summary Получение профиля пользователя
// @Tags profiles
// @Accept json
// @Produce json
// @Success 200 {object} dto.Profile "Profile retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile [get]
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	profile, err := h.profileUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToProfileDTO(profile)

	write.JSONResponse(w, http.StatusOK, response)
}

// UpdateProfile godoc
// @Summary Обновление профиля пользователя
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body dto.Profile true "Profile update data"
// @Success 200 {object} dto.Profile "Profile updated successfully"
// @Failure 400 {object} map[string]string "Bad request - Invalid profile data or missing body"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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

	profile, err := h.profileUsecase.UpdateProfile(r.Context(), userID, *updateProfile)
	if err != nil {
		if errors.Is(err, profiles.ErrInvalidProfileData) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToProfileDTO(profile)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteProfile godoc
// @Summary Удаление профиля пользователя
// @Tags profiles
// @Accept json
// @Produce json
// @Success 204 "Profile deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile [delete]
func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	if err := h.profileUsecase.DeleteProfile(r.Context(), userID); err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	auth.DeleteCookie(w, h.jwtConfig.CookieName, h.jwtConfig.Secure)
	w.WriteHeader(http.StatusNoContent)
}

// GetAvatar godoc
// @Summary Получение аватара профиля
// @Tags avatars
// @Accept json
// @Produce json
// @Success 200 {object} dto.Avatar "Avatar retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 404 {object} map[string]string "Avatar not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile/avatar [get]
func (h *ProfileHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	avatar, err := h.profileUsecase.GetAvatar(r.Context(), userID)
	if err != nil {
		switch err {
		case profiles.ErrAvatarNotFound:
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
// @Summary Загрузка аватара профиля
// @Tags avatars
// @Accept mpfd
// @Produce json
// @Param file formData file true "Avatar image file (JPEG, PNG, or WEBP)"
// @Success 201 {object} dto.Avatar "Avatar uploaded successfully"
// @Failure 400 {object} map[string]string "Bad request - Invalid MIME type"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 413 {object} map[string]string "File too large (max 100MB)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile/avatar [post]
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
	defer file.Close()

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
// @Summary Удаление аватара профиля
// @Tags avatars
// @Accept json
// @Produce json
// @Success 204 "Avatar deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 404 {object} map[string]string "Avatar not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile/avatar [delete]
func (h *ProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	if err := h.profileUsecase.DeleteAvatar(r.Context(), userID); err != nil {
		switch err {
		case profiles.ErrAvatarNotFound:
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangePassword godoc
// @Summary Смена пароля профиля
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body dto.UpdatePassword true "Old and new password data"
// @Success 200 {object} dto.Profile "Password changed successfully"
// @Failure 400 {object} map[string]string "Bad request - Invalid password data or missing body"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid UserID"
// @Failure 404 {object} map[string]string "User not found or wrong password"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /profile/password [put]
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

	var dtoUpdatePassword dto.UpdatePassword
	if err := body.GetBody(r, &dtoUpdatePassword); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidPasswordData)
		return
	}

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	updatedProfile, err := h.profileUsecase.ChangePassword(r.Context(), userID, dtoUpdatePassword.OldPassword, dtoUpdatePassword.NewPassword)
	if err != nil {
		switch err {
		case profiles.ErrUserNotExist:
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case profiles.ErrWrongPassword:
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToProfileDTO(updatedProfile)

	write.JSONResponse(w, http.StatusOK, response)
}
