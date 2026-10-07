package alisa_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"alisa-gpt/internal/models"
	suite_factory "alisa-gpt/internal/pkg/suite/factory"
	suite_provider "alisa-gpt/internal/pkg/suite/provider"
)

const (
	testUserID = "user-1"

	firstQuestion  = "первый вопрос"
	firstAnswer    = "первый ответ"
	secondQuestion = "второй вопрос"
	secondAnswer   = "второй ответ"
)

// newTestService — реальный сервис навыка поверх мока LLM: тестируется
// настоящий сервис, мокается только внешний мир.
func newTestService(t *testing.T) (*suite_provider.Provider, context.Context) {
	t.Helper()

	sp, cleanup := suite_provider.NewProvider()
	t.Cleanup(cleanup)

	sp.GetLLMClientMock().Test(t)

	return sp, sp.Context()
}

func buildRequest(command string, newSession bool) models.AliceRequest {
	return suite_factory.NewRequestFactory().
		WithUserID(testUserID).
		WithCommand(command).
		WithNewSession(newSession).
		Build()
}

func TestHandleNewSessionWelcome(t *testing.T) {
	sp, ctx := newTestService(t)

	resp := sp.GetAlisaService().Handle(ctx, buildRequest("привет", true))

	require.Equal(t, sp.GetConfig().Dialog.Welcome, resp.Response.Text)
	require.False(t, resp.Response.EndSession)
	require.Equal(t, "1.0", resp.Version)
}

func TestHandleExitCommandEndsSession(t *testing.T) {
	sp, ctx := newTestService(t)

	resp := sp.GetAlisaService().Handle(ctx, buildRequest("стоп", false))

	require.True(t, resp.Response.EndSession)
	require.NotEmpty(t, resp.Response.Text)
}

func TestHandleResetClearsHistory(t *testing.T) {
	sp, ctx := newTestService(t)

	sp.GetHistory().Append(testUserID, models.Message{Role: models.RoleUser, Content: "старый вопрос"})
	sp.GetHistory().Append(testUserID, models.Message{Role: models.RoleAssistant, Content: "старый ответ"})

	resp := sp.GetAlisaService().Handle(ctx, buildRequest("новый диалог", false))
	require.Equal(t, "Начинаем новый диалог. Спроси что-нибудь.", resp.Response.Text)

	// после сброса LLM видит только системный промпт + новую реплику
	sp.GetLLMClientMock().
		EXPECT().
		Chat(mock.Anything, []models.Message{
			{Role: models.RoleSystem, Content: sp.GetConfig().OpenAI.SystemPrompt},
			{Role: models.RoleUser, Content: "новый вопрос"},
		}).
		Return("новый ответ", nil).
		Once()

	answer := sp.GetAlisaService().Handle(ctx, buildRequest("новый вопрос", false))
	require.Equal(t, "новый ответ", answer.Response.Text)
}

func TestHandleAnswerKeepsHistoryRoles(t *testing.T) {
	sp, ctx := newTestService(t)
	svc := sp.GetAlisaService()

	sp.GetLLMClientMock().
		EXPECT().
		Chat(mock.Anything, []models.Message{
			{Role: models.RoleSystem, Content: sp.GetConfig().OpenAI.SystemPrompt},
			{Role: models.RoleUser, Content: firstQuestion},
		}).
		Return(firstAnswer, nil).
		Once()

	first := svc.Handle(ctx, buildRequest(firstQuestion, false))
	require.Equal(t, firstAnswer, first.Response.Text)

	// во второй реплике история уже содержит роли user и assistant
	sp.GetLLMClientMock().
		EXPECT().
		Chat(mock.Anything, []models.Message{
			{Role: models.RoleSystem, Content: sp.GetConfig().OpenAI.SystemPrompt},
			{Role: models.RoleUser, Content: firstQuestion},
			{Role: models.RoleAssistant, Content: firstAnswer},
			{Role: models.RoleUser, Content: secondQuestion},
		}).
		Return(secondAnswer, nil).
		Once()

	second := svc.Handle(ctx, buildRequest(secondQuestion, false))
	require.Equal(t, secondAnswer, second.Response.Text)
}

func TestHandleLLMErrorReturnsFallback(t *testing.T) {
	sp, ctx := newTestService(t)

	sp.GetLLMClientMock().
		EXPECT().
		Chat(mock.Anything, mock.Anything).
		Return("", errors.New("boom")).
		Once()

	resp := sp.GetAlisaService().Handle(ctx, buildRequest("сложный вопрос", false))

	require.Equal(t, sp.GetConfig().Dialog.Fallback, resp.Response.Text)
	require.False(t, resp.Response.EndSession)
}

func TestHandleTruncatesLongAnswer(t *testing.T) {
	sp, ctx := newTestService(t)

	longReply := strings.Repeat("а", 80)

	sp.GetLLMClientMock().
		EXPECT().
		Chat(mock.Anything, mock.Anything).
		Return(longReply, nil).
		Once()

	resp := sp.GetAlisaService().Handle(ctx, buildRequest("вопрос", false))

	maxChars := sp.GetConfig().Dialog.MaxChars
	require.LessOrEqual(t, len([]rune(resp.Response.Text)), maxChars)
	require.True(t, strings.HasSuffix(resp.Response.Text, "…"))
}

func TestHandleSanitizesMarkdownInTTS(t *testing.T) {
	sp, ctx := newTestService(t)

	sp.GetLLMClientMock().
		EXPECT().
		Chat(mock.Anything, mock.Anything).
		Return("Жирный *текст* и `код`", nil).
		Once()

	resp := sp.GetAlisaService().Handle(ctx, buildRequest("вопрос", false))

	require.Equal(t, "Жирный *текст* и `код`", resp.Response.Text)
	require.Equal(t, "Жирный текст и код", resp.Response.TTS)
}
