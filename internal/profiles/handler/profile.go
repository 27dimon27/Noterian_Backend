package handler

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
}

type ProfileHandler struct {
	profileUsecase ProfileUsecase
}

func NewProfileHandler(profileUsecase ProfileUsecase) *ProfileHandler {
	return &ProfileHandler{
		profileUsecase: profileUsecase,
	}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(string)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, profiles.ErrInvalidUserID)
		return
	}

	profile, err := h.profileUsecase.GetProfile(r.Context(), userUUID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ProfileResponse{
		Profile: profile,
	}

	write.JSONResponse(w, http.StatusOK, response)
}
