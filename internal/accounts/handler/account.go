package handler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

type AccountHandler struct {
	accountUsecase usecase.AccountUsecase
}

type AccountResponse struct {
	Account *models.Account `json:"account"`
}

func NewAccountHandler(accountUsecase usecase.AccountUsecase) *AccountHandler {
	return &AccountHandler{
		accountUsecase: accountUsecase,
	}
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(string)
	if !ok {
		helpers.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, accounts.ErrInvalidUserID)
		return
	}

	account, err := h.accountUsecase.GetAccount(userUUID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := AccountResponse{
		Account: account,
	}

	helpers.JSONResponse(w, http.StatusOK, response)
}
