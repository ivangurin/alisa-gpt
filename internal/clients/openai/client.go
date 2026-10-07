// Package openai — клиент OpenAI Chat Completions поверх net/http:
// bearer-токен, модель/лимиты из конфига, таймаут через ctx, маппинг
// ошибок API в сентинелы llm.ErrTimeout / llm.ErrUnavailable.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"alisa-gpt/internal/clients/llm"
	config_pkg "alisa-gpt/internal/config"
	"alisa-gpt/internal/models"
)

const (
	chatCompletionsPath = "/chat/completions"

	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "

	contentTypeJSON = "application/json"
)

// Client — реализация llm.Client для OpenAI-совместимого API.
type Client struct {
	http *http.Client

	baseURL   string
	model     string
	maxTokens int
	timeoutMS int
	apiKey    string
}

var _ llm.Client = (*Client)(nil)

// New создаёт клиент из секции openai конфига.
func New(cfg config_pkg.OpenAIConfig) *Client {
	return &Client{
		// транспортный таймаут чуть больше прикладного: основной механизм
		// ограничения — ctx в Chat
		http:      &http.Client{Timeout: msToDuration(cfg.TimeoutMS + transportMarginMS)},
		baseURL:   cfg.BaseURL,
		model:     cfg.Model,
		maxTokens: cfg.MaxTokens,
		timeoutMS: cfg.TimeoutMS,
		apiKey:    cfg.APIKey,
	}
}

// Chat отправляет историю в Chat Completions и возвращает текст ответа.
// Ошибки классифицируются в сентинелы llm (см. errors_classify.go).
func (c *Client) Chat(ctx context.Context, messages []models.Message) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, msToDuration(c.timeoutMS))
	defer cancel()

	payload := chatRequest{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: c.maxTokens,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode request: %w: %w", err, llm.ErrUnavailable)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+chatCompletionsPath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w: %w", err, llm.ErrUnavailable)
	}

	req.Header.Set(authorizationHeader, bearerPrefix+c.apiKey)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", classifyRequestError(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", classifyStatusError(resp)
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode response: %w: %w", err, llm.ErrUnavailable)
	}

	if len(decoded.Choices) == 0 || decoded.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("%w: empty choices", llm.ErrUnavailable)
	}

	return decoded.Choices[0].Message.Content, nil
}
