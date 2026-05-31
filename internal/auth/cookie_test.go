package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteCookie(t *testing.T) {
	t.Run("clears cookie with secure=false", func(t *testing.T) {
		w := httptest.NewRecorder()
		DeleteCookie(w, "session", false)

		resp := w.Result()
		cookies := resp.Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected exactly 1 cookie, got %d", len(cookies))
		}

		c := cookies[0]
		if c.Name != "session" {
			t.Errorf("expected name=session, got %q", c.Name)
		}
		if c.Value != "" {
			t.Errorf("expected empty value, got %q", c.Value)
		}
		if !c.HttpOnly {
			t.Error("expected HttpOnly=true")
		}
		if c.Secure {
			t.Error("expected Secure=false")
		}
		if c.SameSite != http.SameSiteStrictMode {
			t.Errorf("expected SameSiteStrict, got %v", c.SameSite)
		}
		if c.MaxAge != -1 {
			t.Errorf("expected MaxAge=-1, got %d", c.MaxAge)
		}
		if c.Path != "/" {
			t.Errorf("expected Path=/, got %q", c.Path)
		}
	})

	t.Run("clears cookie with secure=true", func(t *testing.T) {
		w := httptest.NewRecorder()
		DeleteCookie(w, "auth", true)

		resp := w.Result()
		cookies := resp.Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected 1 cookie, got %d", len(cookies))
		}
		if !cookies[0].Secure {
			t.Error("expected Secure=true")
		}
	})
}
