package alisa

import (
	"context"

	"alisa-gpt/internal/models"
)

// answer — основной путь: история + реплика → LLM → ответ. Ошибка LLM
// (включая таймаут) отдаёт fallback-фразу и сохраняет сессию живой:
// ретраев внутри одного запроса нет — лимит вебхука 4.5 с не оставляет
// времени на повтор.
func (s *Service) answer(ctx context.Context, req models.AliceRequest, command string) models.AliceResponse {
	messages := s.chatMessages(req.Session.UserID, command)

	reply, err := s.llm.Chat(ctx, messages)
	if err != nil {
		s.log.Warn(ctx, "llm call failed, replying with fallback", map[string]any{
			fieldCommand: command,
			fieldUserID:  req.Session.UserID,
			fieldError:   err.Error(),
		})

		return s.respond(req, s.config.Dialog.Fallback, false)
	}

	s.history.Append(req.Session.UserID, models.Message{Role: models.RoleUser, Content: command})
	s.history.Append(req.Session.UserID, models.Message{Role: models.RoleAssistant, Content: reply})

	text := truncate(reply, s.config.Dialog.MaxChars)

	s.log.Info(ctx, "answer sent", map[string]any{
		fieldCommand: command,
		fieldUserID:  req.Session.UserID,
		fieldChars:   len(text),
	})

	return s.respond(req, text, false)
}

// chatMessages собирает запрос к LLM: системный промпт + история диалога
// + текущая реплика пользователя (ещё не сохранённая в историю).
func (s *Service) chatMessages(userID, command string) []models.Message {
	messages := make([]models.Message, 0, s.config.History.Limit+2)
	messages = append(messages, models.Message{Role: models.RoleSystem, Content: s.config.OpenAI.SystemPrompt})
	messages = append(messages, s.history.Snapshot(userID)...)
	messages = append(messages, models.Message{Role: models.RoleUser, Content: command})

	return messages
}
