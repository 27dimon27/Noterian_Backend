package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
)

func TestLogger(t *testing.T) {
	origLogger := slog.Default()
	defer slog.SetDefault(origLogger)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	t.Run("logs request start and stop", func(t *testing.T) {
		buf.Reset()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			requestID := r.Context().Value(types.RequestIDKey)
			if requestID == nil {
				t.Errorf("expected request_id in context")
			}
			if _, ok := requestID.(string); !ok {
				t.Errorf("expected request_id to be string, got %T", requestID)
			}

			w.WriteHeader(http.StatusOK)
		})

		handler := Logger(next)

		req := httptest.NewRequest("GET", "/path", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if !nextCalled {
			t.Errorf("expected next handler to be called")
		}

		logLines := bytes.Split(buf.Bytes(), []byte("\n"))
		if len(logLines) > 0 && len(logLines[len(logLines)-1]) == 0 {
			logLines = logLines[:len(logLines)-1]
		}

		if len(logLines) != 2 {
			t.Errorf("expected 2 log lines, got %d", len(logLines))
		}
	})

	t.Run("logs correct request data", func(t *testing.T) {
		buf.Reset()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})

		handler := Logger(next)

		req := httptest.NewRequest("POST", "/api/users", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		logOutput := buf.String()

		if !contains(logOutput, "request started") {
			t.Errorf("expected 'request started' in logs")
		}
		if !contains(logOutput, "request completed") {
			t.Errorf("expected 'request completed' in logs")
		}
		if !contains(logOutput, "POST") {
			t.Errorf("expected method POST in logs")
		}
		if !contains(logOutput, "/api/users") {
			t.Errorf("expected path /api/users in logs")
		}
		if !contains(logOutput, "201") {
			t.Errorf("expected status 201 in logs")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
