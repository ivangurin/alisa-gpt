package models

// AliceRequest — входящий запрос вебхука Яндекс.Диалогов (протокол v1.0).
// Поля, которые навык не использует (nlu, payload, state и пр.), не читаются.
type AliceRequest struct {
	Meta    AliceMeta        `json:"meta"`
	Session AliceSession     `json:"session"`
	Request AliceRequestBody `json:"request"`
	Version string           `json:"version"`
}

// AliceMeta — метаданные запроса (локаль, таймзона, источник вызова).
type AliceMeta struct {
	Locale   string `json:"locale"`
	Timezone string `json:"timezone"`
	ClientID string `json:"client_id"`
}

// AliceSession — идентификаторы сессии диалога. UserID стабилен между
// запусками навыка (ключ истории диалога), SessionID меняется на каждый запуск.
type AliceSession struct {
	MessageID int64  `json:"message_id"`
	SessionID string `json:"session_id"`
	SkillID   string `json:"skill_id"`
	UserID    string `json:"user_id"`
	New       bool   `json:"new"`
}

// AliceRequestBody — реплика пользователя. Command — нормализованный текст
// (строчные, без пунктуации), OriginalUtterance — как распознала речь.
type AliceRequestBody struct {
	Command           string `json:"command"`
	OriginalUtterance string `json:"original_utterance"`
	Type              string `json:"type"` // SimpleUtterance | ButtonPressed
}

// AliceResponse — ответ вебхука по протоколу v1.0.
type AliceResponse struct {
	Response AliceResponseBody `json:"response"`
	Session  AliceSession      `json:"session"`
	Version  string            `json:"version"`
}

// AliceResponseBody — текст ответа. TTS озвучивается, Text показывается
// в приложении; лимит Алисы на Text — 1024 символа.
type AliceResponseBody struct {
	Text       string `json:"text"`
	TTS        string `json:"tts"`
	EndSession bool   `json:"end_session"`
}
