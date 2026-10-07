package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"alisa-gpt/internal/clients/llm"
	"alisa-gpt/internal/clients/openai"
	config_pkg "alisa-gpt/internal/config"
	"alisa-gpt/internal/models"
)

func testConfig(baseURL string, timeoutMS int) config_pkg.OpenAIConfig {
	return config_pkg.OpenAIConfig{
		BaseURL:   baseURL,
		Model:     "test-model",
		MaxTokens: 42,
		TimeoutMS: timeoutMS,
		APIKey:    "test-key",
	}
}

func chatMessages() []models.Message {
	return []models.Message{
		{Role: models.RoleSystem, Content: "system prompt"},
		{Role: models.RoleUser, Content: "привет"},
	}
}

func TestChatSuccess(t *testing.T) {
	var gotAuth string
	var gotPayload struct {
		Model     string           `json:"model"`
		Messages  []models.Message `json:"messages"`
		MaxTokens int              `json:"max_tokens"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assert (не require): паника require из хендлера роняет весь тест без отчёта
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)

		gotAuth = r.Header.Get("Authorization")
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"Привет! Чем помочь?"}}]}`))
	}))
	defer srv.Close()

	client := openai.New(testConfig(srv.URL+"/v1", 1000))

	reply, err := client.Chat(context.Background(), chatMessages())
	require.NoError(t, err)
	require.Equal(t, "Привет! Чем помочь?", reply)
	require.Equal(t, "Bearer test-key", gotAuth)
	require.Equal(t, "test-model", gotPayload.Model)
	require.Equal(t, 42, gotPayload.MaxTokens)
	require.Equal(t, chatMessages(), gotPayload.Messages)
}

func TestChatTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	client := openai.New(testConfig(srv.URL, 50))

	_, err := client.Chat(context.Background(), chatMessages())
	require.ErrorIs(t, err, llm.ErrTimeout)
}

func TestChatHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := openai.New(testConfig(srv.URL, 1000))

	_, err := client.Chat(context.Background(), chatMessages())
	require.ErrorIs(t, err, llm.ErrUnavailable)
}

func TestChatEmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	client := openai.New(testConfig(srv.URL, 1000))

	_, err := client.Chat(context.Background(), chatMessages())
	require.ErrorIs(t, err, llm.ErrUnavailable)
	require.NotErrorIs(t, err, llm.ErrTimeout)
}
