package csrf

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler() (*Handler, config.CSRFConfig) {
	cfg := config.CSRFConfig{
		CookieName: "csrf_token",
		CookieTime: 24 * time.Hour,
		Secure:     true,
	}
	handler := NewHandler(cfg)
	return handler, cfg
}

func TestNewHandler(t *testing.T) {
	cfg := config.CSRFConfig{
		CookieName: "test_cookie",
		CookieTime: 1 * time.Hour,
		Secure:     false,
	}

	handler := NewHandler(cfg)
	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.cfg)
}

func TestHandler_GetToken(t *testing.T) {
	t.Run("successfully returns token", func(t *testing.T) {
		handler, cfg := setupTestHandler()

		req := httptest.NewRequest("GET", "/csrf-token", nil)
		w := httptest.NewRecorder()

		handler.GetToken(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check cookie
		cookies := resp.Cookies()
		require.Len(t, cookies, 1)
		cookie := cookies[0]
		assert.Equal(t, cfg.CookieName, cookie.Name)
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, "/", cookie.Path)

		// Check response body
		var response TokenResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.CSRFToken)

		// Validate that the token in response matches the hashed token in cookie
		hashedTokenFromCookie := cookie.Value
		assert.True(t, Validate(response.CSRFToken, hashedTokenFromCookie))
	})

	t.Run("returns unique tokens on each request", func(t *testing.T) {
		handler, _ := setupTestHandler()

		var tokens []string

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/csrf-token", nil)
			w := httptest.NewRecorder()

			handler.GetToken(w, req)

			var response TokenResponse
			err := json.NewDecoder(w.Result().Body).Decode(&response)
			require.NoError(t, err)

			tokens = append(tokens, response.CSRFToken)
		}

		// Check that all tokens are unique
		uniqueTokens := make(map[string]bool)
		for _, token := range tokens {
			uniqueTokens[token] = true
		}
		assert.Equal(t, len(tokens), len(uniqueTokens))
	})

	t.Run("sets cookie with correct config values", func(t *testing.T) {
		cfg := config.CSRFConfig{
			CookieName: "custom_csrf",
			CookieTime: 12 * time.Hour,
			Secure:     false,
		}
		handler := NewHandler(cfg)

		req := httptest.NewRequest("GET", "/csrf-token", nil)
		w := httptest.NewRecorder()

		handler.GetToken(w, req)

		cookie := w.Result().Cookies()[0]
		assert.Equal(t, "custom_csrf", cookie.Name)
		assert.Equal(t, int(12*time.Hour.Seconds()), cookie.MaxAge)
		assert.False(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		assert.True(t, cookie.HttpOnly)
	})

	t.Run("response content type is application/json", func(t *testing.T) {
		handler, _ := setupTestHandler()

		req := httptest.NewRequest("GET", "/csrf-token", nil)
		w := httptest.NewRecorder()

		handler.GetToken(w, req)

		contentType := w.Result().Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json")
	})

	t.Run("response has correct JSON structure", func(t *testing.T) {
		handler, _ := setupTestHandler()

		req := httptest.NewRequest("GET", "/csrf-token", nil)
		w := httptest.NewRecorder()

		handler.GetToken(w, req)

		var response map[string]interface{}
		err := json.NewDecoder(w.Result().Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "csrf_token")
		token, ok := response["csrf_token"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, token)
	})

	t.Run("handles concurrent requests", func(t *testing.T) {
		handler, _ := setupTestHandler()

		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/csrf-token", nil)
				w := httptest.NewRecorder()

				handler.GetToken(w, req)

				assert.Equal(t, http.StatusOK, w.Result().StatusCode)
				done <- true
			}()
		}

		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}

// Integration test example
func TestCSRFFlow(t *testing.T) {
	handler, cfg := setupTestHandler()

	// Step 1: Get CSRF token
	req1 := httptest.NewRequest("GET", "/csrf-token", nil)
	w1 := httptest.NewRecorder()
	handler.GetToken(w1, req1)

	resp1 := w1.Result()
	defer resp1.Body.Close()

	// Extract token from response
	var tokenResp TokenResponse
	err := json.NewDecoder(resp1.Body).Decode(&tokenResp)
	require.NoError(t, err)

	// Extract cookie
	cookies := resp1.Cookies()
	require.Len(t, cookies, 1)

	// Step 2: Use token in subsequent request
	req2 := httptest.NewRequest("POST", "/protected", nil)
	req2.Header.Set("X-CSRF-Token", tokenResp.CSRFToken)
	req2.AddCookie(cookies[0])

	// Validate token
	cookieToken, err := GetFromCookie(req2, cfg)
	require.NoError(t, err)

	isValid := Validate(tokenResp.CSRFToken, cookieToken)
	assert.True(t, isValid)
}
