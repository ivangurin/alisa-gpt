package models

// Источники ошибок API.
const (
	ErrorSourceInternal = "INTERNAL"
)

// ErrorDetail — машиночитаемая часть ошибки API.
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

// ErrorResponse — единый контракт ошибок API.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
