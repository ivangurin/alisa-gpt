package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"alisa-gpt/internal/clients/llm"
)

// maxErrorBodyBytes — сколько байт тела ошибки API прочитать для лога/обёртки.
const maxErrorBodyBytes = 1024

// classifyRequestError маппит транспортную ошибку в сентинелы llm:
// дедлайн ctx (включая отмену вебхуком) — ErrTimeout, прочее — ErrUnavailable.
func classifyRequestError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("llm request: %w: %w", err, llm.ErrTimeout)
	}

	return fmt.Errorf("llm request: %w: %w", err, llm.ErrUnavailable)
}

// classifyStatusError маппит неуспешный HTTP-статус в llm.ErrUnavailable,
// прикладывая фрагмент тела ответа (для диагностики в логах).
func classifyStatusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes)) //nolint:errcheck // тело нужно только для текста ошибки

	return fmt.Errorf("llm status %d: %s: %w", resp.StatusCode, string(body), llm.ErrUnavailable)
}
