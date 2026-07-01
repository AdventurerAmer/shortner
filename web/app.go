package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
)

type App struct {
	cfg    *config.ServiceConfig
	logger *logging.Logger
}

func New(cfg *config.ServiceConfig, logger *logging.Logger) *App {
	app := &App{
		cfg:    cfg,
		logger: logger,
	}
	return app
}

func (app *App) Run(router http.Handler) {
	cfg := app.cfg
	logger := app.logger

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Info("http server started", "port", cfg.Port)
		defer logger.Info("http server ended", "port", cfg.Port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	select {
	case err := <-errCh:
		logger.Error("'ListenAndServe' failed", "error", err)
	case <-sigCtx.Done():
		logger.Info("graceful shutdown started")
		defer logger.Info("graceful shutdown ended")

		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed, forcing close", "error", err)
			if err := srv.Close(); err != nil {
				logger.Error("server force close failed", "error", err)
			}
		}
	}
}
