package suite_provider

import (
	"time"

	"alisa-gpt/internal/config"
)

// GetConfig возвращает фиксированный тестовый конфиг: короткий таймаут,
// малый лимит истории, заданные фразы — тесты пинают эти значения.
func (p *Provider) GetConfig() *config.Config {
	if p.config == nil {
		p.config = &config.Config{
			App: config.App{
				Host:        "127.0.0.1",
				Port:        "0",
				HTTPTimeout: 5,
			},
			Log: config.LogConfig{Level: "error", Format: "dev"},
			OpenAI: config.OpenAIConfig{
				BaseURL:      "http://localhost:0/v1",
				Model:        "test-model",
				MaxTokens:    64,
				TimeoutMS:    1000,
				SystemPrompt: "Ты тестовый собеседник. Отвечай кратко.",
				APIKey:       "test-api-key",
			},
			Dialog: config.DialogConfig{
				Welcome:  "Привет! Я тестовый чат.",
				Fallback: "Тестовый fallback.",
				MaxChars: 50,
			},
			History: config.HistoryConfig{
				Limit: 4,
				TTL:   time.Hour,
			},
		}
	}

	return p.config
}
