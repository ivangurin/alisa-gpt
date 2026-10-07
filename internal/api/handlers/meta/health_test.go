package meta_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"alisa-gpt/internal/api/handlers/base"
	"alisa-gpt/internal/api/handlers/meta"
	logger_pkg "alisa-gpt/internal/pkg/logger"
)

func TestHealthCheck(t *testing.T) {
	handler := meta.NewHandler(base.NewHandler(logger_pkg.NewLogger(nil, "error", "json")))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)

	err := handler.HealthCheck(rec, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, "ok", body["status"])
}
