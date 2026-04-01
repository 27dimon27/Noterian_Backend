package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

func TestAuth(t *testing.T) {
	jwtConfig := config.JWTConfig{
		Secret:        "secret-key",
		CookieName:    "auth",
		CookieTimeJWT: 3600,
		Secure:        false,
	}

	userID := uuid.New()
	token, err := jwt.GenerateToken(userID.String(), jwtConfig.CookieTimeJWT*time.Second, jwtConfig.Secret)
	if err != nil {
		t.Fatalf("failed to generate token: %s", err)
	}

	t.Run("success valid token", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			userIDFromCtx := r.Context().Value(types.UserIDKey)
			if userIDFromCtx == nil {
				t.Errorf("expected UserID in context")
			}

			parsedID, ok := userIDFromCtx.(uuid.UUID)
			if !ok {
				t.Errorf("expected uuid.UUID, got %T", userIDFromCtx)
			}
			if parsedID != userID {
				t.Errorf("expected userID %v, got %v", userID, parsedID)
			}

			w.WriteHeader(http.StatusOK)
		})

		handler := Auth(next, jwtConfig)
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  jwtConfig.CookieName,
			Value: token,
		})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if !nextCalled {
			t.Errorf("expected next handler to be called")
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("missing cookie", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("next handler should not be called")
		})

		handler := Auth(next, jwtConfig)
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("next handler should not be called")
		})

		handler := Auth(next, jwtConfig)
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  jwtConfig.CookieName,
			Value: "invalid-token",
		})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
}
