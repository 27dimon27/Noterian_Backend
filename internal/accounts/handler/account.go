package handler

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

type AccountUsecase interface {
	GetAccount(ctx context.Context, userID uuid.UUID) (*models.Account, error)
}

type AccountHandler struct {
	accountUsecase AccountUsecase
}

func NewAccountHandler(accountUsecase AccountUsecase) *AccountHandler {
	return &AccountHandler{
		accountUsecase: accountUsecase,
	}
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(string)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, accounts.ErrInvalidUserID)
		return
	}

	account, err := h.accountUsecase.GetAccount(r.Context(), userUUID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.AccountResponse{
		Account: account,
	}

	write.JSONResponse(w, http.StatusOK, response)
}
