// Package llm определяет контракт клиента языковой модели: потребитель
// (сервис навыка) знает интерфейс, реализация живёт в internal/clients/openai.
package llm

import (
	"context"

	"alisa-gpt/internal/models"
)

// Client — синхронный запрос чата: полная история (system + реплики) на вход,
// текст ответа модели на выход. Реализация обязана уважать ctx (таймаут
// запроса ограничен лимитом Алисы на ответ вебхука).
type Client interface {
	Chat(ctx context.Context, messages []models.Message) (string, error)
}
