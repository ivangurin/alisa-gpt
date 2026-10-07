// Package middleware — глобальные middleware роутера: request/response-логи.
package middleware

import (
	"net/http"
	"time"

	logger_pkg "alisa-gpt/internal/pkg/logger"
	"alisa-gpt/internal/pkg/metadata"
)

// field-константы для логов.
const (
	fieldMethod     = "method"
	fieldPath       = "path"
	fieldStatus     = "status"
	fieldDurationMS = "duration_ms"
)

// Logging логирует каждый запрос: метод, путь, статус, длительность.
// request_id попадает в запись автоматически — логгер читает его из ctx
// (chi middleware.RequestID → metadata.GetRequestID).
func Logging(log logger_pkg.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.Info(r.Context(), "http request", map[string]any{
				fieldMethod:     r.Method,
				fieldPath:       r.URL.Path,
				fieldStatus:     rec.status,
				fieldDurationMS: time.Since(started).Milliseconds(),
				"request_id":    metadata.GetRequestID(r.Context()),
			})
		})
	}
}

// statusRecorder запоминает статус ответа для лога.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
