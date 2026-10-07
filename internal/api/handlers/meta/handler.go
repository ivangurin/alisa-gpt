// Package meta — мета-ручки приложения: живость и healthcheck.
package meta

import (
	"alisa-gpt/internal/api/handlers/base"
)

// Handler обрабатывает запросы к мета-информации приложения.
type Handler struct {
	*base.Handler
}

// NewHandler создаёт новый Handler.
func NewHandler(baseHandler *base.Handler) *Handler {
	return &Handler{Handler: baseHandler}
}
