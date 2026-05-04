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
		ctx = context.WithValue(ctx, types.JWTTokenKey, cookieJWT.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
