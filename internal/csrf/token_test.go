package csrf

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		token1, err := Generate()
		require.NoError(t, err)
		assert.NotEmpty(t, token1)

		token2, err := Generate()
		require.NoError(t, err)
		assert.NotEmpty(t, token2)

		// Tokens should be different
		assert.NotEqual(t, token1, token2)
	})

	t.Run("generates valid base64", func(t *testing.T) {
		token, err := Generate()
		require.NoError(t, err)

		// Should be valid base64 URL encoding
		_, err = base64.URLEncoding.DecodeString(token)
		assert.NoError(t, err)
	})

	t.Run("generates correct length", func(t *testing.T) {
		token, err := Generate()
		require.NoError(t, err)

		decoded, err := base64.URLEncoding.DecodeString(token)
		require.NoError(t, err)
		assert.Equal(t, tokenLength, len(decoded))
	})
}

func TestHashToken(t *testing.T) {
	t.Run("same token produces same hash", func(t *testing.T) {
		token := "test-token"
		hash1 := hashToken(token)
		hash2 := hashToken(token)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		hash1 := hashToken("token1")
		hash2 := hashToken("token2")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("empty token produces valid hash", func(t *testing.T) {
		hash := hashToken("")
		assert.NotEmpty(t, hash)
	})

	t.Run("hash is deterministic", func(t *testing.T) {
		token := "consistent-token"
		expected := hashToken(token)
		for i := 0; i < 10; i++ {
			assert.Equal(t, expected, hashToken(token))
		}
	})
}

func TestSetCookie(t *testing.T) {
	cfg := config.CSRFConfig{
		CookieName: "csrf_token",
		CookieTime: 24 * time.Hour,
		Secure:     true,
	}

	t.Run("sets cookie with correct attributes", func(t *testing.T) {
		w := httptest.NewRecorder()
		token := "test-token"

		SetCookie(w, token, cfg)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, cfg.CookieName, cookie.Name)
		assert.Equal(t, hashToken(token), cookie.Value)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, int(cfg.CookieTime.Seconds()), cookie.MaxAge)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	})

	t.Run("sets cookie with Secure=false", func(t *testing.T) {
		cfgInsecure := config.CSRFConfig{
			CookieName: "csrf_token",
			CookieTime: 1 * time.Hour,
			Secure:     false,
		}

		w := httptest.NewRecorder()
		SetCookie(w, "token", cfgInsecure)

		cookie := w.Result().Cookies()[0]
		assert.False(t, cookie.Secure)
	})

	t.Run("multiple cookies don't interfere", func(t *testing.T) {
		w := httptest.NewRecorder()

		SetCookie(w, "token1", cfg)
		SetCookie(w, "token2", cfg)

		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 2)
	})
}

func TestGetFromCookie(t *testing.T) {
	cfg := config.CSRFConfig{
		CookieName: "csrf_token",
		CookieTime: 24 * time.Hour,
		Secure:     true,
	}

	t.Run("returns cookie value when present", func(t *testing.T) {
		expectedValue := "hashed-token-value"
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{
			Name:  cfg.CookieName,
			Value: expectedValue,
		})

		value, err := GetFromCookie(r, cfg)
		require.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	})

	t.Run("returns error when cookie missing", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)

		value, err := GetFromCookie(r, cfg)
		assert.Error(t, err)
		assert.Empty(t, value)
		assert.Equal(t, http.ErrNoCookie, err)
	})

	t.Run("returns error for wrong cookie name", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{
			Name:  "different_name",
			Value: "some-value",
		})

		value, err := GetFromCookie(r, cfg)
		assert.Error(t, err)
		assert.Empty(t, value)
	})

	t.Run("handles empty cookie value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{
			Name:  cfg.CookieName,
			Value: "",
		})

		value, err := GetFromCookie(r, cfg)
		require.NoError(t, err)
		assert.Empty(t, value)
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid token pair returns true", func(t *testing.T) {
		requestToken := "valid-token"
		cookieToken := hashToken(requestToken)

		assert.True(t, Validate(requestToken, cookieToken))
	})

	t.Run("invalid token pair returns false", func(t *testing.T) {
		requestToken := "token1"
		cookieToken := hashToken("token2")

		assert.False(t, Validate(requestToken, cookieToken))
	})

	t.Run("empty request token returns false", func(t *testing.T) {
		cookieToken := hashToken("some-token")

		assert.False(t, Validate("", cookieToken))
	})

	t.Run("empty cookie token returns false", func(t *testing.T) {
		requestToken := "some-token"

		assert.False(t, Validate(requestToken, ""))
	})

	t.Run("both tokens empty returns false", func(t *testing.T) {
		assert.False(t, Validate("", ""))
	})

	t.Run("case sensitivity matters", func(t *testing.T) {
		requestToken := "Token"
		cookieToken := hashToken("token")

		assert.False(t, Validate(requestToken, cookieToken))
	})

	t.Run("special characters validation", func(t *testing.T) {
		requestToken := "token-with-special-!@#$%^&*()"
		cookieToken := hashToken(requestToken)

		assert.True(t, Validate(requestToken, cookieToken))
	})
}
