package worker

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
)

type App struct {
	Config
	logger   *logging.Logger
	consumer ports.Consumer
}

func New(logger *logging.Logger, consumer ports.Consumer, opts ...Option) (*App, error) {
	cfg := Config{
		gracefulShutdownTimeout: 10 * time.Second,
		ackTimeout:              time.Second,
		ackRetries:              10,
	}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	app := &App{
		Config:   cfg,
		logger:   logger,
		consumer: consumer,
	}
	return app, nil
}

func (app *App) Run(cfg *config.Config, handler ports.ConsumerHandlerFunc) error {
	logger := app.logger

	// TODO: hardcoding version here
	shutdown, err := telemetry.New(cfg, "Worker", "0.0.1")
	if err != nil {
		return fmt.Errorf("'telemetry.New' failed: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		shutdown(ctx)
	}()

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh := app.consumer.Receive(sigCtx)
	done := make(chan struct{})

	go func() {
		for msg := range msgCh {
			if err := handler(ctx, msg); err != nil {
				logger.Error("handle message failed", "error", err)
				continue
			}
			var lastErr error
			for range app.ackRetries {
				err := func() error {
					dctx, cancel := context.WithTimeout(ctx, app.ackTimeout)
					defer cancel()
					if err := app.consumer.Ack(dctx, msg); err != nil {
						return fmt.Errorf("'consumer.Ack' failed: %w", err)
					}
					return nil
				}()
				if err != nil {
					lastErr = err
				} else {
					lastErr = nil
					break
				}
			}
			if lastErr != nil {
				logger.Error("ack message failed", "error", lastErr)
			} else {
				logger.Debug("acked message successfully")
			}
		}
		close(done)
	}()

	<-sigCtx.Done()

	select {
	case <-done:
	case <-time.After(app.gracefulShutdownTimeout):
		return fmt.Errorf("graceful shutdown timeout")
	}

	return nil
}
