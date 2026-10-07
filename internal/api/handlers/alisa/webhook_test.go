package alisa_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	alisa_handler "alisa-gpt/internal/api/handlers/alisa"
	"alisa-gpt/internal/api/handlers/base"
	"alisa-gpt/internal/models"
	logger_pkg "alisa-gpt/internal/pkg/logger"
)

// stubService — детерминированная замена сервиса навыка.
type stubService struct {
	gotRequest models.AliceRequest
}

const testVersion = "1.0"

func (s *stubService) Handle(_ context.Context, req models.AliceRequest) models.AliceResponse {
	s.gotRequest = req

	return models.AliceResponse{
		Response: models.AliceResponseBody{Text: "ответ сервиса"},
		Version:  testVersion,
	}
}

func newWebhookHandler(t *testing.T) (*alisa_handler.Handler, *stubService) {
	t.Helper()

	svc := &stubService{}
	handler := alisa_handler.NewHandler(base.NewHandler(logger_pkg.NewLogger(nil, "error", "json")), svc)

	return handler, svc
}

func TestWebhookHappyPath(t *testing.T) {
	handler, svc := newWebhookHandler(t)

	body, err := json.Marshal(models.AliceRequest{Version: testVersion})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	require.NoError(t, handler.Webhook(rec, req))
	require.Equal(t, http.StatusOK, rec.Code)

	var decoded models.AliceResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&decoded))
	require.Equal(t, "ответ сервиса", decoded.Response.Text)
	require.Equal(t, testVersion, decoded.Version)
	require.Equal(t, testVersion, svc.gotRequest.Version)
}

func TestWebhookMalformedBody(t *testing.T) {
	handler, _ := newWebhookHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not json")))

	require.NoError(t, handler.Webhook(rec, req))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	var decoded models.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&decoded))
	require.Equal(t, http.StatusBadRequest, decoded.Error.Code)
}
