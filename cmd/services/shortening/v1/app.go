package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	logging.New(cfg)

	logging.Info("config was loaded successfully", "environment", cfg.App.Environment, "version", cfg.App.Version)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", *cfg.Services.Shortening.Port),
		MaxHeaderBytes:    *cfg.Services.Shortening.MaxHeaderBytes,
		ReadHeaderTimeout: *cfg.Services.Shortening.ReadHeaderTimeout,
		ReadTimeout:       *cfg.Services.Shortening.ReadTimeout,
		WriteTimeout:      *cfg.Services.Shortening.WriteTimeout,
		IdleTimeout:       *cfg.Services.Shortening.IdelTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		logging.Info("http server started", "service", *cfg.Services.Shortening.Name, "port", cfg.Services.Shortening.Port)
		defer logging.Info("http server ended", "service", *cfg.Services.Shortening.Name, "port", cfg.Services.Shortening.Port)

		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}
	}()

	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	select {
	case err := <-errCh:
		logging.Error("listen and serve failed", "error", err, "service", *cfg.Services.Shortening.Name)
	case <-signalCtx.Done():
		logging.Info("graceful shutdown started", "service", *cfg.Services.Shortening.Name)
		defer logging.Info("graceful shutdown ended", "service", *cfg.Services.Shortening.Name)

		ctx, cancel := context.WithTimeout(context.Background(), *cfg.Services.Shortening.GracefulShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logging.Error("graceful shutdown failed, forcing close", "error", err, "service", *cfg.Services.Shortening.Name)
			if err := srv.Close(); err != nil {
				logging.Error("server force close failed", "error", err, "service", *cfg.Services.Shortening.Name)
			}
		}
	}

	return 0
}
