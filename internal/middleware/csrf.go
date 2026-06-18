package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
)

func CSRF(next http.Handler, csrfConfig config.CSRFConfig, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieToken, err := csrf.GetFromCookie(r, csrfConfig)
		if err != nil {
			logger.Warn(fmt.Sprintf("[%s:%s] %v", "middleware", "CSRF", csrf.ErrCSRFTokenMissing))
			write.JSONErrorResponse(w, http.StatusForbidden, csrf.PublicMsgErrCSRFTokenMissing)
			return
		}

		requestToken := r.Header.Get(csrfConfig.HeaderName)
		if requestToken == "" {
			logger.Warn(fmt.Sprintf("[%s:%s] %v", "middleware", "CSRF", csrf.ErrCSRFTokenMissing))
			write.JSONErrorResponse(w, http.StatusForbidden, csrf.PublicMsgErrCSRFTokenMissing)
			return
		}

		if !csrf.Validate(requestToken, cookieToken) {
			logger.Warn(fmt.Sprintf("[%s:%s] %v", "middleware", "CSRF", csrf.ErrCSRFTokenInvalid))
			write.JSONErrorResponse(w, http.StatusForbidden, csrf.PublicMsgErrCSRFTokenInvalid)
			return
		}

		next.ServeHTTP(w, r)
	})
}
