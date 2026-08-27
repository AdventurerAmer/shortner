package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
)

type ReadinessHandler func(ctx context.Context) error

type App struct {
	*config.ServiceConfig
	cfg          *config.Config
	logger       *logging.Logger
	ready        atomic.Bool
	shuttingDown atomic.Bool
}

func New(serviceCfg *config.ServiceConfig, cfg *config.Config, logger *logging.Logger) *App {
	app := &App{
		ServiceConfig: serviceCfg,
		cfg:           cfg,
		logger:        logger,
	}
	return app
}

func (app *App) Run(mux *Mux, readiness ReadinessHandler) error {
	logger := app.logger

	shutdown, err := telemetry.New(app.cfg, app.Name, app.Version)
	if err != nil {
		return fmt.Errorf("'telemetry.New' failed: %w", err)
	}

	livez := func(w http.ResponseWriter, r *http.Request) {
		// Pure liveness: process is running.
		// Optional: add a very cheap deadlock detector or goroutine count guard.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
	mux.serveMux.HandleFunc("/livez", livez)

	readyz := func(w http.ResponseWriter, r *http.Request) {
		if app.shuttingDown.Load() || !app.ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		if err := readiness(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "not_ready",
				"checks": map[string]string{"database": err.Error()},
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
	mux.serveMux.HandleFunc("/readyz", readyz)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		shutdown(ctx)
	}()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Port),
		Handler:           mux,
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

	app.ready.Store(true)

	select {
	case err := <-errCh:
		return err
	case <-sigCtx.Done():
		app.shuttingDown.Store(true)

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
