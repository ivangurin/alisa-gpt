// Package config читает конфигурацию навыка из окружения: дефолты заданы
// envDefault-тегами, после парсинга применяется fail-fast валидация.
// Секрет (ключ OpenAI) дефолта не имеет. Локально переменные можно
// держать в .env — файл подгружается перед парсингом.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

// Дефолты длинных текстов: многострочные фразы в envDefault-теге нечитаемы,
// поэтому задаются константами и применяются к пустым значениям после парсинга.
const (
	DefaultOpenAISystemPrompt = "Ты — голосовой собеседник внутри колонки. " +
		"Отвечай кратко и по делу, двумя-четырьмя предложениями, живым разговорным языком. " +
		"Не используй markdown, списки, эмодзи и разметку: ответ озвучивается синтезатором речи."

	DefaultWelcomeText  = "Привет! Я чат с GPT. Спроси меня о чём угодно."
	DefaultFallbackText = "Не успел подумать. Спроси ещё раз или попроще."
)

const (
	dotEnvFile = ".env"

	maxTextLen   = 1024 // лимит протокола Алисы на response.text
	minTimeoutMS = 500
	maxTimeoutMS = 4000 // лимит Алисы на ответ вебхука — 4.5 c
)

var (
	ErrBadPort         = errors.New("app port must be in 1..65535")
	ErrBadHTTPTimeout  = errors.New("http timeout must be positive")
	ErrBadTimeoutMS    = errors.New("openai timeout-ms must be in 500..4000")
	ErrBadMaxTokens    = errors.New("openai max-tokens must be positive")
	ErrBadMaxChars     = errors.New("dialog max-chars must be in 100..1024")
	ErrBadHistoryLimit = errors.New("history limit must be positive")
	ErrBadHistoryTTL   = errors.New("history ttl must be positive")
	ErrNoAPIKey        = errors.New("openai api key is not set")
)

// App — адрес и таймауты HTTP-сервера.
type App struct {
	Host        string `env:"APP_HOST" envDefault:"0.0.0.0"`
	Port        string `env:"APP_PORT" envDefault:"8080"`
	HTTPTimeout int    `env:"HTTP_TIMEOUT" envDefault:"30"`
}

// LogConfig — настройки логгера.
type LogConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"` // debug | info | warn | error
	Format string `env:"LOG_FORMAT" envDefault:"dev"` // dev | text | json
}

// OpenAIConfig — параметры вызова OpenAI Chat Completions. TimeoutMS держит
// запас против лимита Алисы на ответ вебхука (4.5 с). APIKey — только из env.
type OpenAIConfig struct {
	BaseURL      string `env:"OPENAI_BASE_URL" envDefault:"https://api.openai.com/v1"`
	Model        string `env:"OPENAI_MODEL" envDefault:"gpt-4o"`
	MaxTokens    int    `env:"OPENAI_MAX_TOKENS" envDefault:"300"`
	TimeoutMS    int    `env:"OPENAI_TIMEOUT_MS" envDefault:"3500"`
	SystemPrompt string `env:"OPENAI_SYSTEM_PROMPT"`

	APIKey string `env:"OPENAI_API_KEY"`
}

// DialogConfig — фразы и ограничения ответа навыка.
type DialogConfig struct {
	Welcome  string `env:"DIALOG_WELCOME"`
	Fallback string `env:"DIALOG_FALLBACK"`
	MaxChars int    `env:"DIALOG_MAX_CHARS" envDefault:"1000"`
}

// HistoryConfig — параметры in-memory истории диалога.
type HistoryConfig struct {
	Limit int           `env:"HISTORY_LIMIT" envDefault:"20"`
	TTL   time.Duration `env:"HISTORY_TTL" envDefault:"30m"`
}

// Config — конфигурация навыка.
type Config struct {
	App     App           `env:""`
	Log     LogConfig     `env:""`
	OpenAI  OpenAIConfig  `env:""`
	Dialog  DialogConfig  `env:""`
	History HistoryConfig `env:""`
}

// NewConfigFromEnv подгружает опциональный .env, парсит окружение
// (envDefault применяются к отсутствующим переменным) и валидирует конфиг.
func NewConfigFromEnv() (*Config, error) {
	if _, statErr := os.Stat(dotEnvFile); statErr == nil {
		if loadErr := godotenv.Load(dotEnvFile); loadErr != nil {
			return nil, fmt.Errorf("load %s: %w", dotEnvFile, loadErr)
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	cfg.applyTextDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyTextDefaults заполняет длинные текстовые дефолты (см. константы).
func (c *Config) applyTextDefaults() {
	if c.OpenAI.SystemPrompt == "" {
		c.OpenAI.SystemPrompt = DefaultOpenAISystemPrompt
	}

	if c.Dialog.Welcome == "" {
		c.Dialog.Welcome = DefaultWelcomeText
	}

	if c.Dialog.Fallback == "" {
		c.Dialog.Fallback = DefaultFallbackText
	}
}

// validate — fail-fast проверки: невалидный конфиг должен ронять старт,
// а не превращаться в 500 на каждом запросе.
func (c *Config) validate() error {
	port, err := strconv.Atoi(c.App.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%w: %q", ErrBadPort, c.App.Port)
	}

	if c.App.HTTPTimeout < 1 {
		return fmt.Errorf("%w: %d", ErrBadHTTPTimeout, c.App.HTTPTimeout)
	}

	if c.OpenAI.TimeoutMS < minTimeoutMS || c.OpenAI.TimeoutMS > maxTimeoutMS {
		return fmt.Errorf("%w: %d", ErrBadTimeoutMS, c.OpenAI.TimeoutMS)
	}

	if c.OpenAI.MaxTokens < 1 {
		return fmt.Errorf("%w: %d", ErrBadMaxTokens, c.OpenAI.MaxTokens)
	}

	if c.Dialog.MaxChars < 100 || c.Dialog.MaxChars > maxTextLen {
		return fmt.Errorf("%w: %d", ErrBadMaxChars, c.Dialog.MaxChars)
	}

	if c.History.Limit < 1 {
		return fmt.Errorf("%w: %d", ErrBadHistoryLimit, c.History.Limit)
	}

	if c.History.TTL < 1 {
		return fmt.Errorf("%w: %s", ErrBadHistoryTTL, c.History.TTL)
	}

	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("%w: OPENAI_API_KEY", ErrNoAPIKey)
	}

	return nil
}
