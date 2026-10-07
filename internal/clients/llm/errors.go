package llm

import "errors"

var (
	// ErrTimeout — модель не успела ответить в отведённый таймаут.
	ErrTimeout = errors.New("llm request timeout")

	// ErrUnavailable — транспортная ошибка или неуспешный ответ API:
	// сеть, не-200, битый JSON.
	ErrUnavailable = errors.New("llm service unavailable")
)
