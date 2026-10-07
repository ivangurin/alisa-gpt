package metadata

import (
	"context"

	"github.com/go-chi/chi/v5/middleware"
)

// GetRequestID возвращает request_id из контекста (middleware.RequestID chi).
func GetRequestID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

// WithRequestID кладёт request_id в контекст (ключ chi, чтобы GetReqID из
// middleware и логгер читали одно и то же). Нужен для переноса request_id
// в отвязанные (detached) контексты долгоживущих горутин.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, middleware.RequestIDKey, requestID)
}
