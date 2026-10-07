package suite_provider

import (
	llm_mocks "alisa-gpt/internal/clients/llm/mocks"
)

// GetLLMClientMock возвращает mockery-мок LLM-клиента — сценарий задаётся
// в тесте через .EXPECT(). TestingT подключается в тесте вызовом Test(t).
func (p *Provider) GetLLMClientMock() *llm_mocks.Client {
	if p.llmMock == nil {
		p.llmMock = &llm_mocks.Client{}
	}

	return p.llmMock
}
