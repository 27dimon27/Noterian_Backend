package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

func Auth(next http.Handler, jwtConfig config.JWTConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieJWT, err := r.Cookie(jwtConfig.CookieName)
		if err != nil {
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrUnauthorized)
			return
		}

		tokenPayload, err := jwt.ValidateToken(cookieJWT.Value, jwtConfig.Secret)
		if err != nil {
			write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrInvalidToken)
			return
		}

		userUUID, err := uuid.Parse(tokenPayload.UserID)
		if err != nil {
			write.JSONErrorResponse(w, http.StatusUnauthorized, auth.ErrInvalidUserID)
			return
		}

		ctx := context.WithValue(r.Context(), types.UserIDKey, userUUID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WebSocketAuth(jwtConfig config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookieJWT, err := r.Cookie(jwtConfig.CookieName)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenPayload, err := jwt.ValidateToken(cookieJWT.Value, jwtConfig.Secret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			userUUID, err := uuid.Parse(tokenPayload.UserID)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), types.UserIDKey, userUUID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
