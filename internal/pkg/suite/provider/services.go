package suite_provider

import (
	"alisa-gpt/internal/pkg/history"
	"alisa-gpt/internal/services/alisa"
)

// GetHistory возвращает реальное хранилище диалогов с параметрами тестового
// конфига.
func (p *Provider) GetHistory() *history.Store {
	if p.history == nil {
		p.history = history.NewStore(p.GetConfig().History.Limit, p.GetConfig().History.TTL)
	}

	return p.history
}

// GetAlisaService возвращает РЕАЛЬНЫЙ сервис навыка поверх мока LLM.
func (p *Provider) GetAlisaService() *alisa.Service {
	return alisa.NewService(p.GetLogger(), p.GetConfig(), p.GetLLMClientMock(), p.GetHistory())
}
