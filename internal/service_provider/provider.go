// Package service_provider — DI-провайдер навыка: ленивые синглтоны истории
// и сервиса поверх переданного LLM-клиента (клиентом провайдер не владеет).
package service_provider

import (
	"context"
	"sync"

	"alisa-gpt/internal/clients/llm"
	"alisa-gpt/internal/config"
	"alisa-gpt/internal/pkg/history"
	logger_pkg "alisa-gpt/internal/pkg/logger"
	"alisa-gpt/internal/services/alisa"
)

// Provider — ленивый DI-контейнер продового стека.
type Provider struct {
	ctx       context.Context
	logger    logger_pkg.Logger
	config    *config.Config
	llmClient llm.Client

	mu           sync.Mutex
	history      *history.Store
	alisaService *alisa.Service
}

// NewProvider создаёт провайдер поверх уже собранного LLM-клиента.
func NewProvider(
	ctx context.Context,
	log logger_pkg.Logger,
	cfg *config.Config,
	llmClient llm.Client,
) *Provider {
	return &Provider{
		ctx:       ctx,
		logger:    log,
		config:    cfg,
		llmClient: llmClient,
	}
}

// GetContext возвращает корневой ctx приложения.
func (p *Provider) GetContext() context.Context {
	return p.ctx
}

// GetLogger возвращает логгер приложения.
func (p *Provider) GetLogger() logger_pkg.Logger {
	return p.logger
}

// GetConfig возвращает конфиг приложения.
func (p *Provider) GetConfig() *config.Config {
	return p.config
}

// GetLLMClient возвращает LLM-клиент, переданный при создании провайдера.
func (p *Provider) GetLLMClient() llm.Client {
	return p.llmClient
}
