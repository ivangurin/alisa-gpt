package service_provider

import (
	"alisa-gpt/internal/pkg/history"
	"alisa-gpt/internal/services/alisa"
)

// GetHistory лениво создаёт хранилище диалогов.
func (p *Provider) GetHistory() *history.Store {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.history == nil {
		p.history = history.NewStore(p.config.History.Limit, p.config.History.TTL)
	}

	return p.history
}

// GetAlisaService лениво создаёт сервис навыка. Зависимости берутся ДО
// захвата p.mu (вложенные ленивые геттеры берут тот же лок — deadlock).
func (p *Provider) GetAlisaService() *alisa.Service {
	historyStore := p.GetHistory()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.alisaService == nil {
		p.alisaService = alisa.NewService(p.logger, p.config, p.llmClient, historyStore)
	}

	return p.alisaService
}
