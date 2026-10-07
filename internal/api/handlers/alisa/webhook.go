package alisa

import (
	"encoding/json"
	"net/http"

	"alisa-gpt/internal/models"
)

// maxBodyBytes — лимит тела запроса вебхука: реплика пользователя
// укладывается в единицы килобайт.
const maxBodyBytes = 1 << 20

// Webhook обрабатывает POST / — запрос навыка по протоколу Яндекс.Диалогов v1.0.
//
//	@Summary		Handle Alice webhook
//	@Description	Единственная ручка навыка: принимает запрос вебхука Яндекс.Диалогов (протокол v1.0), зовёт ChatGPT и возвращает ответ для озвучивания. Ошибки LLM не роняют сессию — навык отвечает fallback-фразой.
//	@Tags			alisa
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.AliceRequest	true	"Запрос вебхука Алисы"
//	@Success		200		{object}	models.AliceResponse
//	@Failure		400		{object}	models.ErrorResponse	"Битный JSON запроса"
//	@Failure		500		{object}	models.ErrorResponse
//	@Router			/ [post]
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req models.AliceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponseWithCode(w, http.StatusBadRequest, "malformed request body: "+err.Error())

		return nil //nolint:nilerr // контракт хендлеров: ответ (400) уже записан, ошибка обработана
	}

	resp := h.service.Handle(r.Context(), req)

	h.WriteJSON(w, resp, http.StatusOK)

	return nil
}
