package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	config_pkg "alisa-gpt/internal/config"
)

func TestNewConfigFromEnvAppliesDefaults(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-api-key")

	cfg, err := config_pkg.NewConfigFromEnv()
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0", cfg.App.Host)
	require.Equal(t, "8080", cfg.App.Port)
	require.Equal(t, 30, cfg.App.HTTPTimeout)
	require.Equal(t, "info", cfg.Log.Level)
	require.Equal(t, "dev", cfg.Log.Format)
	require.Equal(t, "https://api.openai.com/v1", cfg.OpenAI.BaseURL)
	require.Equal(t, "gpt-4o", cfg.OpenAI.Model)
	require.Equal(t, 300, cfg.OpenAI.MaxTokens)
	require.Equal(t, 3500, cfg.OpenAI.TimeoutMS)
	require.Equal(t, config_pkg.DefaultOpenAISystemPrompt, cfg.OpenAI.SystemPrompt)
	require.Equal(t, config_pkg.DefaultWelcomeText, cfg.Dialog.Welcome)
	require.Equal(t, config_pkg.DefaultFallbackText, cfg.Dialog.Fallback)
	require.Equal(t, 1000, cfg.Dialog.MaxChars)
	require.Equal(t, 20, cfg.History.Limit)
	require.Equal(t, 30*time.Minute, cfg.History.TTL)
	require.Equal(t, "test-api-key", cfg.OpenAI.APIKey)
}

func TestNewConfigFromEnvOverrides(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-api-key")
	t.Setenv("APP_PORT", "9000")
	t.Setenv("OPENAI_MODEL", "gpt-4o-mini")
	t.Setenv("OPENAI_BASE_URL", "https://proxy.example.com/v1")
	t.Setenv("OPENAI_TIMEOUT_MS", "2000")
	t.Setenv("HISTORY_TTL", "10m")
	t.Setenv("OPENAI_SYSTEM_PROMPT", "Будь краток.")

	cfg, err := config_pkg.NewConfigFromEnv()
	require.NoError(t, err)

	require.Equal(t, "9000", cfg.App.Port)
	require.Equal(t, "gpt-4o-mini", cfg.OpenAI.Model)
	require.Equal(t, "https://proxy.example.com/v1", cfg.OpenAI.BaseURL)
	require.Equal(t, 2000, cfg.OpenAI.TimeoutMS)
	require.Equal(t, 10*time.Minute, cfg.History.TTL)
	require.Equal(t, "Будь краток.", cfg.OpenAI.SystemPrompt)
}

func TestNewConfigFromEnvRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	_, err := config_pkg.NewConfigFromEnv()
	require.ErrorIs(t, err, config_pkg.ErrNoAPIKey)
}

func TestNewConfigFromEnvRejectsBadValues(t *testing.T) {
	tests := []struct {
		name    string
		setEnv  func(t *testing.T)
		wantErr error
	}{
		{
			name: "bad port",
			setEnv: func(t *testing.T) {
				t.Setenv("APP_PORT", "70000")
			},
			wantErr: config_pkg.ErrBadPort,
		},
		{
			name: "timeout over alice limit",
			setEnv: func(t *testing.T) {
				t.Setenv("OPENAI_TIMEOUT_MS", "4500")
			},
			wantErr: config_pkg.ErrBadTimeoutMS,
		},
		{
			name: "max chars over protocol limit",
			setEnv: func(t *testing.T) {
				t.Setenv("DIALOG_MAX_CHARS", "2000")
			},
			wantErr: config_pkg.ErrBadMaxChars,
		},
		{
			name: "negative history limit",
			setEnv: func(t *testing.T) {
				t.Setenv("HISTORY_LIMIT", "-1")
			},
			wantErr: config_pkg.ErrBadHistoryLimit,
		},
		{
			name: "zero http timeout",
			setEnv: func(t *testing.T) {
				t.Setenv("HTTP_TIMEOUT", "0")
			},
			wantErr: config_pkg.ErrBadHTTPTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OPENAI_API_KEY", "test-api-key")
			tt.setEnv(t)

			_, err := config_pkg.NewConfigFromEnv()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
