package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
)

var currentLogger *slog.Logger

func Init() *slog.Logger {
	logLevel := os.Getenv("LOG_LEVEL")

	var level slog.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	currentLogger = logger

	return logger
}

func WithRequest(ctx context.Context, args ...any) (*slog.Logger, []any) {
	if requestID, ok := ctx.Value(types.RequestIDKey).(string); ok {
		args = append(args, "request_id", requestID)
	}
	if currentLogger == nil {
		currentLogger = slog.Default()
	}
	return currentLogger, args
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func Debug(ctx context.Context, msg string, args ...any) {
	log, updatedArgs := WithRequest(ctx, args...)
	log.DebugContext(ctx, msg, updatedArgs...)
}

func Info(ctx context.Context, msg string, args ...any) {
	log, updatedArgs := WithRequest(ctx, args...)
	log.InfoContext(ctx, msg, updatedArgs...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	log, updatedArgs := WithRequest(ctx, args...)
	log.WarnContext(ctx, msg, updatedArgs...)
}

func Error(ctx context.Context, msg string, args ...any) {
	log, updatedArgs := WithRequest(ctx, args...)
	updatedArgs = append(updatedArgs, "stacktrace", getStackTrace())
	log.ErrorContext(ctx, msg, updatedArgs...)
}
