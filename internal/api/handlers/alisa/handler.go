// Package alisa — вебхук навыка Яндекс.Алисы: единственная ручка протокола
// v1.0, разбирающая запрос и отдающая ответ через сервис навыка.
package alisa

import (
	"context"

	"alisa-gpt/internal/api/handlers/base"
	"alisa-gpt/internal/models"
)

// Service — потребляемый сервис навыка. Интерфейс объявлен у потребителя
// (в этом пакете), реализация — *services/alisa.Service.
type Service interface {
	Handle(ctx context.Context, req models.AliceRequest) models.AliceResponse
}

// Handler обрабатывает запросы вебхука Алисы.
type Handler struct {
	*base.Handler
	service Service
}

// NewHandler создаёт новый Handler.
func NewHandler(baseHandler *base.Handler, service Service) *Handler {
	return &Handler{Handler: baseHandler, service: service}
}
