// Package suite_factory — gofakeit-билдеры фикстур для тестов навыка.
package suite_factory

import (
	"github.com/brianvoe/gofakeit/v7"

	"alisa-gpt/internal/models"
)

// RequestFactory строит запросы вебхука Алисы с дефолтами из gofakeit.
type RequestFactory struct {
	userID     string
	sessionID  string
	command    string
	newSession bool
}

// NewRequestFactory создаёт фабрику со случайными дефолтами: случайные
// идентификаторы, случайная реплика, сессия продолжается.
func NewRequestFactory() *RequestFactory {
	return &RequestFactory{
		userID:     gofakeit.UUID(),
		sessionID:  gofakeit.UUID(),
		command:    gofakeit.Sentence(5),
		newSession: false,
	}
}

// WithUserID задаёт user_id (ключ истории диалога).
func (f *RequestFactory) WithUserID(userID string) *RequestFactory {
	f.userID = userID

	return f
}

// WithSessionID задаёт session_id текущего запуска навыка.
func (f *RequestFactory) WithSessionID(sessionID string) *RequestFactory {
	f.sessionID = sessionID

	return f
}

// WithCommand задаёт нормализованную реплику пользователя (command).
func (f *RequestFactory) WithCommand(command string) *RequestFactory {
	f.command = command

	return f
}

// WithNewSession помечает запрос как первое сообщение сессии.
func (f *RequestFactory) WithNewSession(newSession bool) *RequestFactory {
	f.newSession = newSession

	return f
}

// Build собирает запрос вебхука.
func (f *RequestFactory) Build() models.AliceRequest {
	return models.AliceRequest{
		Meta: models.AliceMeta{
			Locale:   "ru-RU",
			Timezone: "Europe/Moscow",
			ClientID: gofakeit.UUID(),
		},
		Session: models.AliceSession{
			MessageID: int64(gofakeit.Number(1, 1000)),
			SessionID: f.sessionID,
			UserID:    f.userID,
			New:       f.newSession,
		},
		Request: models.AliceRequestBody{
			Command:           f.command,
			OriginalUtterance: f.command,
			Type:              "SimpleUtterance",
		},
		Version: "1.0",
	}
}
