package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
)

func CSRF(next http.Handler, csrfConfig config.CSRFConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieToken, err := csrf.GetFromCookie(r, csrfConfig)
		if err != nil {
			write.JSONErrorResponse(w, http.StatusForbidden, csrf.ErrCSRFTokenMissing)
			return
		}

		requestToken := r.Header.Get(csrfConfig.HeaderName)
		if requestToken == "" {
			write.JSONErrorResponse(w, http.StatusForbidden, csrf.ErrCSRFTokenMissing)
			return
		}

		if !csrf.Validate(requestToken, cookieToken) {
			write.JSONErrorResponse(w, http.StatusForbidden, csrf.ErrCSRFTokenInvalid)
			return
		}

		next.ServeHTTP(w, r)
	})
}
