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
	cfg        *config.Config
	serviceCfg *config.ServiceConfig
	logger     *logging.Logger
}

func New(cfg *config.Config, serviceCfg *config.ServiceConfig, logger *logging.Logger) *App {
	app := &App{
		cfg:        cfg,
		serviceCfg: serviceCfg,
		logger:     logger,
	}
	return app
}

func (app *App) Run(router http.Handler) error {
	cfg := app.serviceCfg
	logger := app.logger

	shutdown, err := telemetry.New(app.cfg, app.serviceCfg.Name, "0.0.1")
	if err != nil {
		return fmt.Errorf("'telemetry.New' failed: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		shutdown(ctx)
	}()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		logger.Info("http server started", "port", cfg.Port)
		defer logger.Info("http server ended", "port", cfg.Port)

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

		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			if err := srv.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
