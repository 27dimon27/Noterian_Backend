package middleware

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type websocketResponseWriter struct {
	http.ResponseWriter
	statusCode int
	hijacked   bool
}

func (rw *websocketResponseWriter) WriteHeader(code int) {
	if rw.statusCode == 0 {
		rw.statusCode = code
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *websocketResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		rw.hijacked = true
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.NewString()

		ctx := context.WithValue(r.Context(), types.RequestIDKey, requestID)
		r = r.WithContext(ctx)

		isWebSocket := r.Header.Get("Upgrade") == "websocket"

		logger.Info(ctx, "request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"is_websocket", isWebSocket,
		)

		if isWebSocket {
			wsRW := &websocketResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wsRW, r)

			if !wsRW.hijacked {
				logger.Error(ctx, "websocket upgrade failed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", wsRW.statusCode,
					"duration_ms", time.Since(start).Milliseconds(),
				)
			} else {
				logger.Info(ctx, "websocket connection established",
					"method", r.Method,
					"path", r.URL.Path,
					"duration_ms", time.Since(start).Milliseconds(),
				)
			}
			return
		}

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		if rw.statusCode >= 500 {
			logger.Error(ctx, "request completed with error",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		} else {
			logger.Info(ctx, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}
	})
}
