// Package base — общий хендлер API-групп: запись JSON-ответов и ошибок
// в едином контракте (models.ErrorResponse).
package base

import (
	"context"
	"encoding/json"
	"net/http"

	"alisa-gpt/internal/models"
	logger_pkg "alisa-gpt/internal/pkg/logger"
)

const contentTypeJSON = "application/json"

// Handler — базовый хендлер с логгером и respond-хелперами.
type Handler struct {
	logger logger_pkg.Logger
}

// NewHandler создаёт базовый хендлер.
func NewHandler(logger logger_pkg.Logger) *Handler {
	return &Handler{logger: logger}
}

// Logger возвращает логгер хендлера.
func (h *Handler) Logger() logger_pkg.Logger {
	return h.logger
}

// WriteJSON пишет data с statusCode; маршализация до заголовков, чтобы
// ошибка кодирования не оставила запрос с уже отправленным 200.
func (h *Handler) WriteJSON(w http.ResponseWriter, data any, statusCode int) {
	body, err := json.Marshal(data)
	if err != nil {
		h.WriteErrorResponseWithCode(w, http.StatusInternalServerError, "failed to encode response")

		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)

	_, _ = w.Write(body)
}

// WriteErrorResponseWithCode отвечает ошибкой единого контракта.
func (h *Handler) WriteErrorResponseWithCode(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)

	resp := models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    statusCode,
			Message: message,
			Source:  models.ErrorSourceInternal,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error(context.Background(), "failed to write error response", map[string]any{"error": err.Error()})
	}
}
