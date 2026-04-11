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

	tests := []struct {
		name         string
		logLevel     string
		expectDebug  bool
		expectInfo   bool
		expectWarn   bool
		expectError  bool
		expectSource bool
	}{
		{
			name:         "default level info when not set",
			logLevel:     "",
			expectDebug:  false,
			expectInfo:   true,
			expectWarn:   true,
			expectError:  true,
			expectSource: false,
		},
		{
			name:         "debug level",
			logLevel:     "DEBUG",
			expectDebug:  true,
			expectInfo:   true,
			expectWarn:   true,
			expectError:  true,
			expectSource: true,
		},
		{
			name:         "info level",
			logLevel:     "INFO",
			expectDebug:  false,
			expectInfo:   true,
			expectWarn:   true,
			expectError:  true,
			expectSource: false,
		},
		{
			name:         "warn level",
			logLevel:     "WARN",
			expectDebug:  false,
			expectInfo:   false,
			expectWarn:   true,
			expectError:  true,
			expectSource: false,
		},
		{
			name:         "error level",
			logLevel:     "ERROR",
			expectDebug:  false,
			expectInfo:   false,
			expectWarn:   false,
			expectError:  true,
			expectSource: false,
		},
		{
			name:         "invalid level defaults to info",
			logLevel:     "INVALID",
			expectDebug:  false,
			expectInfo:   true,
			expectWarn:   true,
			expectError:  true,
			expectSource: false,
		},
		{
			name:         "lowercase level works",
			logLevel:     "debug",
			expectDebug:  true,
			expectInfo:   true,
			expectWarn:   true,
			expectError:  true,
			expectSource: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
			}

			origStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			logger := Init()

			w.Close()
			os.Stdout = origStdout

			if logger == nil {
				t.Errorf("expected non-nil logger")
			}

			ctx := context.Background()
			if logger.Enabled(ctx, slog.LevelDebug) != tt.expectDebug {
				t.Errorf("Debug level enabled: got %v, want %v", logger.Enabled(ctx, slog.LevelDebug), tt.expectDebug)
			}
			if logger.Enabled(ctx, slog.LevelInfo) != tt.expectInfo {
				t.Errorf("Info level enabled: got %v, want %v", logger.Enabled(ctx, slog.LevelInfo), tt.expectInfo)
			}
			if logger.Enabled(ctx, slog.LevelWarn) != tt.expectWarn {
				t.Errorf("Warn level enabled: got %v, want %v", logger.Enabled(ctx, slog.LevelWarn), tt.expectWarn)
			}
			if logger.Enabled(ctx, slog.LevelError) != tt.expectError {
				t.Errorf("Error level enabled: got %v, want %v", logger.Enabled(ctx, slog.LevelError), tt.expectError)
			}
		})
	}
}

func TestLogFunctions(t *testing.T) {
	origLogger := currentLogger
	defer func() { currentLogger = origLogger }()

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	currentLogger = logger
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

	t.Run("Debug without request_id", func(t *testing.T) {
		buf.Reset()
		ctx := context.Background()

		Debug(ctx, "debug message no id", "key2", "value2")

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		if logEntry["msg"] != "debug message no id" {
			t.Errorf("expected msg 'debug message no id', got %v", logEntry["msg"])
		}
		if _, ok := logEntry["request_id"]; ok {
			t.Errorf("expected no request_id in log")
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

	t.Run("Info with request_id", func(t *testing.T) {
		buf.Reset()
		ctx := context.WithValue(context.Background(), types.RequestIDKey, "req-789")

		Info(ctx, "info with id", "status", "ok")

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		if logEntry["request_id"] != "req-789" {
			t.Errorf("expected request_id req-789, got %v", logEntry["request_id"])
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

	t.Run("Error logs with stacktrace", func(t *testing.T) {
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
		if _, ok := logEntry["stacktrace"]; !ok {
			t.Errorf("expected stacktrace field in error log")
		}
	})

	t.Run("Error with request_id", func(t *testing.T) {
		buf.Reset()
		ctx := context.WithValue(context.Background(), types.RequestIDKey, "error-req-456")

		Error(ctx, "error with id", "details", "crash")

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		if logEntry["request_id"] != "error-req-456" {
			t.Errorf("expected request_id error-req-456, got %v", logEntry["request_id"])
		}
		if _, ok := logEntry["stacktrace"]; !ok {
			t.Errorf("expected stacktrace field in error log")
		}
	})
}
