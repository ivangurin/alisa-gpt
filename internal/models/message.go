package models

// Роли сообщений чата (совместимы с OpenAI Chat Completions).
const (
	RoleSystem    string = "system"
	RoleUser      string = "user"
	RoleAssistant string = "assistant"
)

// Message — сообщение чата: роль + текст. Используется и как элемент
// истории диалога, и как элемент запроса к LLM (json-теги совпадают
// с форматом OpenAI).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
