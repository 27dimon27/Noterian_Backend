package handler

import (
	"bytes"
	"context"
	"errors"
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
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	profile, err := h.profileUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, profiles.ErrUserNotExist) {
			h.logger.Warn("User not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToProfileDTO(*profile)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in UpdateProfile", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	var dtoUpdateProfile dto.Profile

	if err := body.GetBody(r, &dtoUpdateProfile); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidProfileData)
		return
	}

	updateProfile := dto.FromProfileDTO(dtoUpdateProfile)

	profile, err := h.profileUsecase.UpdateProfile(r.Context(), userID, updateProfile)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrInvalidProfileData), errors.Is(err, profiles.ErrUsernameExists), errors.Is(err, profiles.ErrUserNotExist):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToProfileDTO(*profile)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	err := h.profileUsecase.DeleteProfile(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrUserNotExist):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
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

func (h *ProfileHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	avatar, err := h.profileUsecase.GetAvatar(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrAvatarNotFound):
			h.logger.Warn("Avatar not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToAvatarDTO(*avatar)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, profiles.MAX_FILE_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			h.logger.Warn("Too large file for avatar")
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, profiles.ErrFileTooLarge)
		} else {
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error("Failed to close file in UploadAvatar", "error", err)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	mimeType := http.DetectContentType(buffer)

	if !profiles.AllowedMimeTypes[mimeType] {
		h.logger.Warn("Invalid MIME-type of file")
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidMimeType)
		return
	}

	avatar, err := h.profileUsecase.UploadAvatar(r.Context(), userID, fileHeader.Filename, fileHeader.Size, mimeType, fileToUpload)
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
	}

	response := dto.ToAvatarDTO(*avatar)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *ProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	if err := h.profileUsecase.DeleteAvatar(r.Context(), userID); err != nil {
		switch {
		case errors.Is(err, profiles.ErrAvatarNotFound):
			h.logger.Warn("Avatar not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in ChangePassword", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, profiles.ErrInvalidUserID)
		return
	}

	var dtoUpdatePassword dto.UpdatePassword

	if err := body.GetBody(r, &dtoUpdatePassword); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidPasswordData)
		return
	}

	updatedProfile, err := h.profileUsecase.ChangePassword(r.Context(), userID, dtoUpdatePassword.OldPassword, dtoUpdatePassword.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrUserNotExist):
			h.logger.Warn("User not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, profiles.ErrWrongPassword), errors.Is(err, profiles.ErrInvalidPasswordData):
			h.logger.Warn("Wrong credentials")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToProfileDTO(*updatedProfile)

	write.JSONResponse(w, http.StatusOK, response)
}
