// Package handlers собирает роутер API: пути объявлены константами, хендлеры
// имеют сигнатуру func(w, r) error и оборачиваются в NewHandler.
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	alisa_handler "alisa-gpt/internal/api/handlers/alisa"
	"alisa-gpt/internal/api/handlers/base"
	"alisa-gpt/internal/api/handlers/meta"
	"alisa-gpt/internal/api/middleware"
	logger_pkg "alisa-gpt/internal/pkg/logger"
	"alisa-gpt/internal/service_provider"
)

// Пути API.
const (
	// Swagger — Swagger UI.
	Swagger = "/swagger/*"

	// HealthCheck — живость сервиса.
	HealthCheck = "/healthcheck"

	// Webhook — вебхук навыка Яндекс.Диалогов (протокол v1.0).
	Webhook = "/"
)

// HandlerFunc — контракт хендлера: сам пишет ответ и возвращает nil;
// не-nil ошибка означает баг в хендлере.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handler — адаптер HandlerFunc в http.Handler.
type Handler struct {
	logger      logger_pkg.Logger
	handlerFunc HandlerFunc
}

// ServeHTTP — safety-net по контракту хендлеров: хендлер сам пишет ответ
// и возвращает nil; не-nil ошибка здесь означает баг в хендлере — отвечаем
// 400, чтобы запрос не остался вовсе без ответа.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.handlerFunc(w, r); err != nil {
		h.logger.Error(r.Context(), "handler returned unexpected error", map[string]any{
			"error": err.Error(),
			"path":  r.URL.Path,
		})

		w.WriteHeader(http.StatusBadRequest)
	}
}

// NewHandler оборачивает HandlerFunc в http.Handler.
func NewHandler(logger logger_pkg.Logger, handlerFunc HandlerFunc) Handler {
	return Handler{
		logger:      logger,
		handlerFunc: handlerFunc,
	}
}

// NewHandlers собирает роутер поверх продового провайдера.
func NewHandlers(sp *service_provider.Provider) chi.Router {
	baseHandler := base.NewHandler(sp.GetLogger())
	metaHandler := meta.NewHandler(baseHandler)
	alisaHandler := alisa_handler.NewHandler(baseHandler, sp.GetAlisaService())

	router := chi.NewRouter()

	// request_id кладётся в ctx (chi RequestID), logging-логгер читает его
	// через metadata; Recoverer гасит паники хендлеров в 500.
	router.Use(chi_middleware.Recoverer)
	router.Use(chi_middleware.RequestID)
	router.Use(middleware.Logging(sp.GetLogger().With("component", "http-logging")))

	router.Get(Swagger, httpSwagger.WrapHandler)

	router.Method(http.MethodGet, HealthCheck, NewHandler(sp.GetLogger(), metaHandler.HealthCheck))
	router.Method(http.MethodPost, Webhook, NewHandler(sp.GetLogger(), alisaHandler.Webhook))

	return router
}
