package handler

import (
	"context"
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

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
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

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
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
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
		return
	}

	updateProfile := dto.FromProfileDTO(dtoUpdateProfile)

	profile, err := h.profileUsecase.UpdateProfile(r.Context(), userID, *updateProfile)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToProfileDTO(profile)

	write.JSONResponse(w, http.StatusOK, response)
}

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

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, auth.ErrMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var dtoUpdatePassword dto.UpdatePassword
	if err := body.GetBody(r, &dtoUpdatePassword); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, auth.ErrInvalidInput)
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
