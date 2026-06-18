package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
)

type App struct {
	Env    config.Env
	Cfg    *config.ServiceConfig
	Logger *logging.Logger
}

func New(env config.Env, logger *logging.Logger, cfg *config.ServiceConfig) *App {
	app := &App{
		Env:    env,
		Cfg:    cfg,
		Logger: logger.With(slog.String("service", cfg.Name)),
	}
	return app
}

func (app *App) Run(router http.Handler) {
	cfg := app.Cfg
	logger := app.Logger

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdelTimeout,
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
		logger.Error("ListenAndServe failed", "error", err)
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
