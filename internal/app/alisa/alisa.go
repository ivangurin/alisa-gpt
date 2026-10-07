// Package alisa собирает приложение: конфиг → логгер → closer → LLM-клиент →
// провайдер → роутер → HTTP-сервер, и управляет graceful shutdown (closer, LIFO).
package alisa

import (
	"context"
	"fmt"
	"os"

	"alisa-gpt/internal/api/handlers"
	"alisa-gpt/internal/clients/openai"
	"alisa-gpt/internal/config"
	"alisa-gpt/internal/pkg/closer"
	logger_pkg "alisa-gpt/internal/pkg/logger"
	hs "alisa-gpt/internal/pkg/servers/http"
	"alisa-gpt/internal/service_provider"
)

// fieldError — имя поля ошибки в логах.
const fieldError = "error"

type app struct {
	config *config.Config
}

// NewApp создаёт приложение поверх загруженного конфига.
func NewApp(cfg *config.Config) *app {
	return &app{config: cfg}
}

// Run запускает вебхук и блокируется до сигнала завершения / фатальной ошибки.
func (a *app) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger_pkg.NewLogger(os.Stdout, a.config.Log.Level, a.config.Log.Format)

	cl := closer.New(log)
	defer func() {
		_ = cl.Close()
	}()

	llmClient := openai.New(a.config.OpenAI)

	sp := service_provider.NewProvider(ctx, log, a.config, llmClient)

	// http server
	httpServer := hs.NewServer(a.config.App.Host, a.config.App.Port, a.config.App.HTTPTimeout, handlers.NewHandlers(sp))

	// Порядок закрытия. Closer выполняет функции в обратном порядке (LIFO),
	// поэтому регистрируем в порядке, обратном желаемой последовательности:
	// http-stop → cancel. Сначала сервер перестаёт принимать новые запросы и
	// дожидается in-flight (ответы Алисе уходят целиком), затем гасится
	// корневой ctx.
	cl.Add(func() error {
		cancel()

		return nil
	})
	cl.Add(a.stopServerFunc(httpServer, log))

	errCh := make(chan error, 1)
	go func() {
		log.Info(ctx, "start http server", map[string]any{
			"host": a.config.App.Host,
			"port": a.config.App.Port,
		})

		if err := httpServer.Start(); err != nil {
			log.Error(ctx, "http server start failed", map[string]any{fieldError: err.Error()})
			errCh <- err
			cl.Stop()
		}
	}()

	cl.Wait()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	default:
		return nil
	}
}

// stopServerFunc останавливает HTTP-сервер (исполняется первым по LIFO).
func (a *app) stopServerFunc(httpServer hs.Server, log logger_pkg.Logger) func() error {
	return func() error {
		if err := httpServer.Stop(context.WithoutCancel(context.Background())); err != nil {
			log.Error(context.Background(), "failed to stop http server", map[string]any{fieldError: err.Error()})

			return err
		}

		log.Info(context.Background(), "http server stopped successfully")

		return nil
	}
}
