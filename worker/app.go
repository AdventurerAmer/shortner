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
	logger   *logging.Logger
	consumer ports.Consumer
}

func New(logger *logging.Logger, consumer ports.Consumer) *App {
	return &App{
		logger:   logger,
		consumer: consumer,
	}
}

func (app *App) Run(handler ports.ConsumerHandlerFunc) int {
	logger := app.logger

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh, doneCh := app.consumer.Receive(sigCtx)

	go func() {
		for msg := range msgCh {
			if err := handler(ctx, msg); err != nil {
				logger.Error("handle message failed", "error", err)
				continue
			}
			var lastErr error
			for range 10 {
				err := func() error {
					dctx, cancel := context.WithTimeout(ctx, time.Second)
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
	}()

	<-sigCtx.Done()

	select {
	case <-doneCh:
	case <-time.After(10 * time.Second):
	}

	return 0
}
