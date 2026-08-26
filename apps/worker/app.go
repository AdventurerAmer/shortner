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
	*config.WorkerConfig
	cfg      *config.Config
	consumer ports.Consumer
	logger   *logging.Logger
}

func New(workerCfg *config.WorkerConfig, cfg *config.Config, consumer ports.Consumer, logger *logging.Logger) *App {
	app := &App{
		WorkerConfig: workerCfg,
		cfg:          cfg,
		consumer:     consumer,
		logger:       logger,
	}
	return app
}

func (app *App) Run(handler ports.ConsumerHandlerFunc) error {
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

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh := app.consumer.Receive(sigCtx)
	done := make(chan struct{})

	logger.Info("consumer started", "group", app.Group)
	defer logger.Info("consumer ended", "group", app.Group)

	go func() {
		for msg := range msgCh {
			if err := app.handleMsg(ctx, msg, handler); err != nil {
				logger.Error("failed to handle message", "error", err)
			}
		}
		close(done)
	}()

	<-sigCtx.Done()

	select {
	case <-done:
	case <-time.After(app.GracefulShutdownTimeout):
		return fmt.Errorf("graceful shutdown timeout")
	}

	return nil
}

func (app *App) handleMsg(ctx context.Context, msg ports.ConsumerMessage, h ports.ConsumerHandlerFunc) error {
	start := time.Now()

	defer func() {
		latency := time.Since(start).Milliseconds()
		telemetry.RequestsLatency.Record(ctx, latency)
		telemetry.RequestsCounter.Add(ctx, 1)
	}()

	if err := app.consumer.Consume(ctx, msg, h); err != nil {
		return fmt.Errorf("'consumer.Consume' failed: %w", err)
	}

	if err := app.ackMsg(ctx, msg); err != nil {
		return fmt.Errorf("'app.ackMsg' failed: %w", err)
	}

	return nil
}

func (app *App) ackMsg(ctx context.Context, msg ports.ConsumerMessage) error {
	var lastErr error
	for range app.AckRetries {
		err := func() error {
			dctx, cancel := context.WithTimeout(ctx, app.AckTimeout)
			defer cancel()
			if err := app.consumer.Ack(dctx, msg); err != nil {
				return fmt.Errorf("'consumer.Ack' failed: %w", err)
			}
			return nil
		}()
		if err != nil {
			lastErr = err
		} else {
			return nil
		}
	}
	return fmt.Errorf("'consumer.Ack' failed: %w", lastErr)
}
