package middleware

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Используем наш кастомный ResponseWriter
		lw := &loggingResponseWriter{w, http.StatusOK}

		next.ServeHTTP(lw, r)

		logData := map[string]interface{}{
			"status":         lw.status,
			"method":         r.Method,
			"path":           r.URL.Path,
			"query":          r.URL.RawQuery,
			"remote_addr":    r.RemoteAddr,
			"user_agent":     r.UserAgent(),
			"content_length": r.ContentLength,
			"duration_ms":    time.Since(start).Milliseconds(),
		}

		logJSON, err := json.MarshalIndent(logData, "", "  ")
		if err != nil {
			logger.Error("Failed to format log data", map[string]interface{}{"error": err.Error()})
			return
		}

		logger.Info("Request processed:\n"+string(logJSON), nil)
	})
}

func (l *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijack not supported")
	}
	return h.Hijack()
}

// Переопределяем WriteHeader для сохранения статуса
func (l *loggingResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}
