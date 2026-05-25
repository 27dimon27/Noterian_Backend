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

// GetToken godoc
// @Summary      Получение CSRF-токена
// @Description  Возвращает CSRF-токен в теле и Set-Cookie. Используется во всех методах, изменяющих состояние (POST/PUT/DELETE), через заголовок X-CSRF-Token.
// @Tags         csrf
// @Produce      json
// @Success      200  {object}  map[string]string  "CSRF token successfully generated"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /csrf-token [get]
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
