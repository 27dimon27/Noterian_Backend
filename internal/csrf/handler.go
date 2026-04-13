package csrf

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
)

type Handler struct {
	cfg config.CSRFConfig
}

func NewHandler(cfg config.CSRFConfig) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	token, err := Generate()
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	SetCookie(w, token, h.cfg)

	write.JSONResponse(w, http.StatusOK, map[string]string{
		"csrf_token": token,
	})
}
