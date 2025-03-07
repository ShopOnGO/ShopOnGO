package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &WrapperWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapper, r)

		logData := map[string]interface{}{
			"status":         wrapper.StatusCode,
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
