// Package suite_provider — DI-контейнер тестов: собирает тот же стек, что
// и прод-провайдер, но LLM-клиент — mockery-мок, сети нет.
// Все тесты пишутся через NewProvider, вручную зависимости не собираются.
package suite_provider

import (
	"context"
	"os"

	llm_mocks "alisa-gpt/internal/clients/llm/mocks"
	"alisa-gpt/internal/config"
	"alisa-gpt/internal/pkg/history"
	logger_pkg "alisa-gpt/internal/pkg/logger"
)

// Provider — тестовый DI: без локов, тесты однопоточные.
type Provider struct {
	ctx     context.Context
	logger  logger_pkg.Logger
	config  *config.Config
	llmMock *llm_mocks.Client
	history *history.Store
}

// NewProvider возвращает провайдер тестового стека и cleanup.
// cleanup всегда вызывается defer'ом из теста.
func NewProvider() (*Provider, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Provider{ctx: ctx}

	return p, cancel
}

// Context возвращает корневой ctx теста.
func (p *Provider) Context() context.Context {
	return p.ctx
}

// ContextWithValue возвращает дочерний ctx с произвольным значением.
func (p *Provider) ContextWithValue(key, val any) context.Context {
	return context.WithValue(p.ctx, key, val)
}

// GetLogger возвращает реальный логгер уровня error (не шумит в выводе тестов).
func (p *Provider) GetLogger() logger_pkg.Logger {
	if p.logger == nil {
		p.logger = logger_pkg.NewLogger(os.Stdout, "error", "dev")
	}

	return p.logger
}
