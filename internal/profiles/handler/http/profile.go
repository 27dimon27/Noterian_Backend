package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, types.AppErrorInterface)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, types.AppErrorInterface)
	DeleteProfile(ctx context.Context, userID uuid.UUID) types.AppErrorInterface
	GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, types.AppErrorInterface)
	UploadAvatar(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, types.AppErrorInterface)
	DeleteAvatar(ctx context.Context, profileID uuid.UUID) types.AppErrorInterface
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, types.AppErrorInterface)
}

type ProfileHandler struct {
	profileUsecase ProfileUsecase
	jwtConfig      config.JWTConfig
	logger         *slog.Logger
}

func NewProfileHandler(profileUsecase ProfileUsecase, jwtConfig config.JWTConfig, logger *slog.Logger) *ProfileHandler {
	return &ProfileHandler{
		profileUsecase: profileUsecase,
		jwtConfig:      jwtConfig,
		logger:         logger,
	}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetProfile", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	profile, customErr := h.profileUsecase.GetProfile(r.Context(), userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), profiles.ErrUserNotExist):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToProfileDTO(*profile)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateProfile", profiles.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateProfile", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateProfile", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	var dtoUpdateProfile dto.Profile

	if err := body.GetBody(r, &dtoUpdateProfile); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateProfile", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, profiles.PublicMsgErrInternalServer)
		return
	}

	updateProfile := dto.FromProfileDTO(dtoUpdateProfile)

	profile, customErr := h.profileUsecase.UpdateProfile(r.Context(), userID, updateProfile)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), profiles.ErrInvalidProfileData), errors.Is(customErr.Unwrap(), profiles.ErrUsernameExists), errors.Is(customErr.Unwrap(), profiles.ErrUserNotExist):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToProfileDTO(*profile)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteProfile", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	if customErr := h.profileUsecase.DeleteProfile(r.Context(), userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), profiles.ErrUserNotExist):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
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

func (h *ProfileHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetAvatar", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	avatar, customErr := h.profileUsecase.GetAvatar(r.Context(), userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), profiles.ErrAvatarNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToAvatarDTO(*avatar)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, profiles.MAX_FILE_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", profiles.ErrFileTooLarge))
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, profiles.PublicMsgErrFileTooLarge)
		} else {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", err))
			write.JSONErrorResponse(w, http.StatusInternalServerError, profiles.PublicMsgErrInternalServer)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, profiles.PublicMsgErrInternalServer)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", err))
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, profiles.PublicMsgErrInternalServer)
		return
	}

	mimeType := http.DetectContentType(buffer)

	if !profiles.AllowedMimeTypes[mimeType] {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", profiles.ErrInvalidMimeType))
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.PublicMsgErrInvalidMimeType)
		return
	}

	if fileHeader.Size > profiles.MAX_FILE_SIZE {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAvatar", profiles.ErrFileTooLarge))
		write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, profiles.PublicMsgErrFileTooLarge)
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	avatar, customErr := h.profileUsecase.UploadAvatar(r.Context(), userID, fileHeader.Filename, fileHeader.Size, mimeType, fileToUpload)
	if customErr != nil {
		h.logger.Error(customErr.Error())
		write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		return
	}

	response := dto.ToAvatarDTO(*avatar)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *ProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteAvatar", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	if customErr := h.profileUsecase.DeleteAvatar(r.Context(), userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), profiles.ErrAvatarNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "ChangePassword", profiles.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "ChangePassword", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "ChangePassword", profiles.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.PublicMsgErrInvalidUserID)
		return
	}

	var dtoUpdatePassword dto.UpdatePassword

	if err := body.GetBody(r, &dtoUpdatePassword); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "ChangePassword", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, profiles.PublicMsgErrInternalServer)
		return
	}

	updatedProfile, customErr := h.profileUsecase.ChangePassword(r.Context(), userID, dtoUpdatePassword.OldPassword, dtoUpdatePassword.NewPassword)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), profiles.ErrUserNotExist):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), profiles.ErrWrongPassword), errors.Is(customErr.Unwrap(), profiles.ErrInvalidPasswordData):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToProfileDTO(*updatedProfile)

	write.JSONResponse(w, http.StatusOK, response)
}
