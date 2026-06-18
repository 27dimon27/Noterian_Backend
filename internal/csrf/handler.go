package csrf

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
)

type Handler struct {
	cfg    config.CSRFConfig
	logger *slog.Logger
}

type TokenResponse struct {
	CSRFToken string `json:"csrf_token"`
}

func NewHandler(cfg config.CSRFConfig, logger *slog.Logger) *Handler {
	return &Handler{
		cfg:    cfg,
		logger: logger,
	}
}

func (h *Handler) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token, err := Generate()
	if err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "GetCSRFToken", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, PublicMsgErrInternalServer)
		return
	}

	SetCookie(w, token, h.cfg)

	write.JSONResponse(w, http.StatusOK, TokenResponse{
		CSRFToken: token,
	})
}
