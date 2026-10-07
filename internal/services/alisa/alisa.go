// Package alisa — сервисный слой навыка: превращает запрос вебхука Алисы
// в ответ, ведя историю диалога и вызывая LLM с учётом лимита ответа (4.5 с).
package alisa

import (
	"context"

	"alisa-gpt/internal/clients/llm"
	config_pkg "alisa-gpt/internal/config"
	"alisa-gpt/internal/models"
	"alisa-gpt/internal/pkg/history"
	logger_pkg "alisa-gpt/internal/pkg/logger"
)

// field-константы для логов.
const (
	fieldError     = "error"
	fieldCommand   = "command"
	fieldUserID    = "user_id"
	fieldChars     = "chars"
	fieldComponent = "component"
)

// protocolVersion — версия протокола в ответе вебхука.
const protocolVersion = "1.0"

// Service — оркестратор одного запроса Алисы: системный промпт + история +
// реплика пользователя → LLM → текст ответа.
type Service struct {
	log     logger_pkg.Logger
	config  *config_pkg.Config
	llm     llm.Client
	history *history.Store
}

// NewService собирает сервис поверх LLM-клиента и хранилища истории.
func NewService(
	log logger_pkg.Logger,
	cfg *config_pkg.Config,
	llmClient llm.Client,
	historyStore *history.Store,
) *Service {
	return &Service{
		log:     log.With(fieldComponent, "alisa-service"),
		config:  cfg,
		llm:     llmClient,
		history: historyStore,
	}
}

// Handle обрабатывает запрос вебхука и всегда возвращает валидный ответ
// Алисы: ошибки LLM не роняют сессию, а превращаются в fallback-фразу.
func (s *Service) Handle(ctx context.Context, req models.AliceRequest) models.AliceResponse {
	switch {
	case req.Session.New:
		s.log.Info(ctx, "session started", map[string]any{fieldUserID: req.Session.UserID})

		return s.respond(req, s.config.Dialog.Welcome, false)
	case req.Request.Command == "":
		return s.respond(req, s.config.Dialog.Welcome, false)
	}

	command := req.Request.Command

	if isExitCommand(command) {
		s.log.Info(ctx, "session finished", map[string]any{fieldUserID: req.Session.UserID})

		return s.respond(req, farewellText, true)
	}

	if isResetCommand(command) {
		s.history.Reset(req.Session.UserID)
		s.log.Info(ctx, "history reset", map[string]any{fieldUserID: req.Session.UserID})

		return s.respond(req, resetDoneText, false)
	}

	return s.answer(ctx, req, command)
}

// respond собирает ответ вебхука по протоколу v1.0.
func (s *Service) respond(req models.AliceRequest, text string, endSession bool) models.AliceResponse {
	return models.AliceResponse{
		Response: models.AliceResponseBody{
			Text:       text,
			TTS:        sanitizeSpeech(text),
			EndSession: endSession,
		},
		Session: req.Session,
		Version: protocolVersion,
	}
}
