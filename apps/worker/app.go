package worker

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
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

func (app *App) Run(handler ports.ConsumerHandlerFunc) error {
	logger := app.logger

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
