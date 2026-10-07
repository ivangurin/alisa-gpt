// Package http — тонкая обёртка над net/http.Server: собирает сервер из
// адреса/таймаутов и роутера, стартует (Start) и останавливает с бюджетом
// (Stop).
package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Server — жизненный цикл HTTP-сервера.
type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type server struct {
	server *http.Server
}

// NewServer создаёт сервер: host/port строками (из конфига), timeout в
// секундах идёт и в ReadHeaderTimeout, и в ReadTimeout; WriteTimeout не
// ставится (ответы пишутся быстро, а стриминг обрезать нельзя).
func NewServer(host, port string, timeout int, router http.Handler) Server {
	server := &server{
		server: &http.Server{
			Addr:              net.JoinHostPort(host, port),
			ReadHeaderTimeout: time.Duration(timeout) * time.Second,
			ReadTimeout:       time.Duration(timeout) * time.Second,
			Handler:           router,
		},
	}

	return server
}

// Start блокируется до остановки сервера; ErrServerClosed — штатный выход.
func (s *server) Start() error {
	err := s.server.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("start http server: %w", err)
		}
	}

	return nil
}

// Stop останавливает сервер и дожидается in-flight запросов в пределах
// стоп-бюджета поверх приходящего ctx (с запасом против лимита Алисы 4.5 с).
func (s *server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("stop http server: %w", err)
	}

	return nil
}
