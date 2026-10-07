package meta

import (
	"net/http"
)

// HealthCheck живость сервиса.
//
//	@Summary	Health check
//	@Tags		meta
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/healthcheck [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) error {
	h.WriteJSON(w, map[string]string{"status": "ok"}, http.StatusOK)

	return nil
}
