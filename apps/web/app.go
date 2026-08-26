package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
)

type App struct {
	*config.ServiceConfig
	cfg    *config.Config
	logger *logging.Logger
}

func New(serviceCfg *config.ServiceConfig, cfg *config.Config, logger *logging.Logger) *App {
	app := &App{
		ServiceConfig: serviceCfg,
		cfg:           cfg,
		logger:        logger,
	}
	return app
}

func (app *App) Run(router http.Handler) error {
	logger := app.logger

	shutdown, err := telemetry.New(app.cfg, app.Name, app.Version)
	if err != nil {
		return fmt.Errorf("'telemetry.New' failed: %w", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		shutdown(ctx)
	}()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Port),
		Handler:           router,
		MaxHeaderBytes:    app.MaxHeaderBytes,
		ReadHeaderTimeout: app.ReadHeaderTimeout,
		ReadTimeout:       app.ReadTimeout,
		WriteTimeout:      app.WriteTimeout,
		IdleTimeout:       app.IdleTimeout,
	}

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		logger.Info("http server started", "port", app.Port)
		defer logger.Info("http server ended", "port", app.Port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-sigCtx.Done():
		logger.Info("graceful shutdown started")
		defer logger.Info("graceful shutdown ended")

		ctx, cancel := context.WithTimeout(context.Background(), app.GracefulShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed, forcing close", "error", err)
			if err := srv.Close(); err != nil {
				logger.Error("server force close failed", "error", err)
			}
		}
	}

	return nil
}
