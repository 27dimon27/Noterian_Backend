package metrics

import (
	"bufio"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors (4xx and 5xx responses)",
		},
		[]string{"method", "path", "status_code", "error_type"},
	)

	httpHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_hits_total",
			Help: "Total number of HTTP requests (hits)",
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestDurationByURL = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds_by_url",
			Help:    "HTTP request duration in seconds grouped by URL path",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestDurationByMethod = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds_by_method",
			Help:    "HTTP request duration in seconds grouped by HTTP method",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusCode == code {
		return
	}
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func normalizePath(path string) string {
	patterns := map[string]string{
		`/notes/\d+`:                        "/notes/{noteId}",
		`/notes/\d+/subnote`:                "/notes/{noteId}/subnote",
		`/notes/\d+/subnote/\d+`:            "/notes/{noteId}/subnote/{subnoteId}",
		`/notes/\d+/blocks/\d+`:             "/notes/{noteId}/blocks/{blockId}",
		`/notes/\d+/blocks/\d+/content`:     "/notes/{noteId}/blocks/{blockId}/content",
		`/notes/\d+/blocks/\d+/move`:        "/notes/{noteId}/blocks/{blockId}/move",
		`/notes/\d+/blocks/\d+/formatting`:  "/notes/{noteId}/blocks/{blockId}/formatting",
		`/notes/\d+/blocks/\d+/attachments`: "/notes/{noteId}/blocks/{blockId}/attachments",
		`/ws/notes/\d+`:                     "/ws/notes/{noteId}",
	}

	result := path
	for pattern, replacement := range patterns {
		matched, err := regexp.MatchString(pattern, result)
		if err == nil && matched {
			re := regexp.MustCompile(pattern)
			result = re.ReplaceAllString(result, replacement)
		}
	}

	return result
}

func getErrorType(statusCode int) string {
	if statusCode >= 400 && statusCode < 500 {
		return "client_error"
	}
	if statusCode >= 500 && statusCode < 600 {
		return "server_error"
	}
	return ""
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()

		normalizedPath := normalizePath(r.URL.Path)

		statusCode := wrapped.statusCode
		statusCodeStr := http.StatusText(statusCode)
		if statusCodeStr == "" {
			statusCodeStr = "unknown"
		}

		httpHitsTotal.WithLabelValues(r.Method, normalizedPath, statusCodeStr).Inc()

		if statusCode >= 400 {
			errorType := getErrorType(statusCode)
			httpErrorsTotal.WithLabelValues(r.Method, normalizedPath, statusCodeStr, errorType).Inc()
		}

		httpRequestDurationByURL.WithLabelValues(r.Method, normalizedPath).Observe(duration)

		httpRequestDurationByMethod.WithLabelValues(r.Method).Observe(duration)
	})
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
