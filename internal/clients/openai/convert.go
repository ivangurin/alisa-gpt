package openai

import (
	"time"

	"alisa-gpt/internal/models"
)

// transportMarginMS — запас транспортного таймаута http.Client поверх
// прикладного таймаута из конфига, чтобы первым срабатывал ctx в Chat.
const transportMarginMS = 500

const headerContentType = "Content-Type"

// chatRequest — тело запроса Chat Completions. models.Message переиспользуется
// как элемент messages: json-теги совпадают с форматом OpenAI.
type chatRequest struct {
	Model     string           `json:"model"`
	Messages  []models.Message `json:"messages"`
	MaxTokens int              `json:"max_tokens"`
}

// chatResponse — используемая часть ответа Chat Completions.
type chatResponse struct {
	Choices []struct {
		Message models.Message `json:"message"`
	} `json:"choices"`
}

func msToDuration(ms int) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
