package csrf

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
)

type Middleware struct {
	cfg config.CSRFConfig
}

func NewMiddleware(cfg config.CSRFConfig) *Middleware {
	return &Middleware{cfg: cfg}
}

func (m *Middleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieToken, err := GetFromCookie(r, m.cfg)
		if err != nil {
			write.JSONErrorResponse(w, http.StatusForbidden, ErrCSRFTokenMissing)
			return
		}

		requestToken := r.Header.Get(m.cfg.HeaderName)
		if requestToken == "" {
			write.JSONErrorResponse(w, http.StatusForbidden, ErrCSRFTokenMissing)
			return
		}

		if !Validate(requestToken, cookieToken) {
			write.JSONErrorResponse(w, http.StatusForbidden, ErrCSRFTokenInvalid)
			return
		}

		next.ServeHTTP(w, r)
	})
}
