package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/config"
)

type App struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

func NewApp(cfg config.Config, handler http.Handler) *App {
	return &App{
		server: &http.Server{
			Addr:              cfg.HTTP.Addr,
			Handler:           handler,
			ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
		},
		shutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}
}

func (a *App) Run(ctx context.Context) error {
	serveResult := make(chan error, 1)
	go func() { serveResult <- a.server.ListenAndServe() }()

	select {
	case err := <-serveResult:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		_ = a.server.Close()
		return fmt.Errorf("shut down HTTP server: %w", err)
	}
	if err := <-serveResult; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	return nil
}
