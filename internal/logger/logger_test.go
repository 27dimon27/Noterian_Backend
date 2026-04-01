package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
)

func TestInit(t *testing.T) {
	origLogLevel := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origLogLevel)

	t.Run("default level info when level not set", func(t *testing.T) {
		os.Clearenv()
		logger := Init()

		if logger == nil {
			t.Errorf("expected non-nil logger")
		}

		if !logger.Enabled(context.Background(), slog.LevelInfo) {
			t.Errorf("expected Info level to be enabled")
		}
		if logger.Enabled(context.Background(), slog.LevelDebug) {
			t.Errorf("expected Debug level to be disabled")
		}
	})

	t.Run("debug level", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LOG_LEVEL", "DEBUG")
		logger := Init()

		if !logger.Enabled(context.Background(), slog.LevelDebug) {
			t.Errorf("expected Debug level to be enabled")
		}
	})

	t.Run("info level", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LOG_LEVEL", "INFO")
		logger := Init()

		if !logger.Enabled(context.Background(), slog.LevelInfo) {
			t.Errorf("expected Info level to be enabled")
		}
		if logger.Enabled(context.Background(), slog.LevelDebug) {
			t.Errorf("expected Debug level to be disabled")
		}
	})

	t.Run("warn level", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LOG_LEVEL", "WARN")
		logger := Init()

		if !logger.Enabled(context.Background(), slog.LevelWarn) {
			t.Errorf("expected Warn level to be enabled")
		}
		if logger.Enabled(context.Background(), slog.LevelInfo) {
			t.Errorf("expected Info level to be disabled")
		}
	})

	t.Run("error level", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LOG_LEVEL", "ERROR")
		logger := Init()

		if !logger.Enabled(context.Background(), slog.LevelError) {
			t.Errorf("expected Error level to be enabled")
		}
		if logger.Enabled(context.Background(), slog.LevelWarn) {
			t.Errorf("expected Warn level to be disabled")
		}
	})
}

func TestHTTPError(t *testing.T) {
	t.Run("error implements error interface", func(t *testing.T) {
		err := HTTPError{
			Code:    404,
			Message: "not found",
			Details: map[string]string{"id": "123"},
		}

		if err.Error() != "not found" {
			t.Errorf("expected message 'not found', got '%s'", err.Error())
		}
	})

	t.Run("error without details", func(t *testing.T) {
		err := HTTPError{
			Code:    500,
			Message: "internal error",
		}

		if err.Error() != "internal error" {
			t.Errorf("expected message 'internal error', got '%s'", err.Error())
		}
	})
}

func TestLogFunctions(t *testing.T) {
	origLogger := slog.Default()
	defer slog.SetDefault(origLogger)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	t.Run("Debug logs with request_id", func(t *testing.T) {
		buf.Reset()
		ctx := context.WithValue(context.Background(), types.RequestIDKey, "req-123456")

		Debug(ctx, "debug message", "key", "value")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Errorf("failed to parse log: %s", err)
		}

		if logEntry["msg"] != "debug message" {
			t.Errorf("expected msg 'debug message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "DEBUG" {
			t.Errorf("expected level DEBUG, got %v", logEntry["level"])
		}
		if logEntry["key"] != "value" {
			t.Errorf("expected key value, got %v", logEntry["key"])
		}
		if logEntry["request_id"] != "req-123456" {
			t.Errorf("expected request_id req-123456, got %v", logEntry["request_id"])
		}
	})

	t.Run("Info logs", func(t *testing.T) {
		buf.Reset()
		ctx := context.Background()

		Info(ctx, "info message", "user", "testuser")

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		if logEntry["msg"] != "info message" {
			t.Errorf("expected msg 'info message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "INFO" {
			t.Errorf("expected level INFO, got %v", logEntry["level"])
		}
		if logEntry["user"] != "testuser" {
			t.Errorf("expected user testuser, got %v", logEntry["user"])
		}
	})

	t.Run("Warn logs", func(t *testing.T) {
		buf.Reset()
		ctx := context.Background()

		Warn(ctx, "warn message", "code", 400)

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		if logEntry["msg"] != "warn message" {
			t.Errorf("expected msg 'warn message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "WARN" {
			t.Errorf("expected level WARN, got %v", logEntry["level"])
		}
		if logEntry["code"] != float64(400) {
			t.Errorf("expected code 400, got %v", logEntry["code"])
		}
	})

	t.Run("Error logs", func(t *testing.T) {
		buf.Reset()
		ctx := context.Background()

		Error(ctx, "error message", "error", "something went wrong")

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		if logEntry["msg"] != "error message" {
			t.Errorf("expected msg 'error message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "ERROR" {
			t.Errorf("expected level ERROR, got %v", logEntry["level"])
		}
		if logEntry["error"] != "something went wrong" {
			t.Errorf("expected error 'something went wrong', got %v", logEntry["error"])
		}
	})
}
