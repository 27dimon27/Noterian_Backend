package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

func Auth(next http.Handler, jwtConfig config.JWTConfig, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieJWT, err := r.Cookie(jwtConfig.CookieName)
		if err != nil {
			logger.Error(fmt.Sprintf("[%s:%s] %v", "middleware", "Auth", err))
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.PublicMsgErrUnauthorized)
			return
		}

		tokenPayload, err := jwt.ValidateToken(cookieJWT.Value, jwtConfig.Secret)
		if err != nil {
			logger.Error(fmt.Sprintf("[%s:%s] %v", "middleware", "Auth", err))
			write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.PublicMsgErrInvalidToken)
			return
		}

		userUUID, err := uuid.Parse(tokenPayload.UserID)
		if err != nil {
			logger.Error(fmt.Sprintf("[%s:%s] %v", "middleware", "Auth", err))
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.PublicMsgErrInvalidUserID)
			return
		}

		ctx := context.WithValue(r.Context(), types.UserIDKey, userUUID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
