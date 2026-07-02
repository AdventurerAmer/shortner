package async

import (
	"context"
	"log/slog"
	"sync"

	"github.com/AdventurerAmer/shortner/logging"
	"github.com/google/uuid"
)

type GoFunc = func(ctx context.Context)

type Orchestrator struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewOrchestrator(parent context.Context) *Orchestrator {
	ctx, cancel := context.WithCancel(parent)
	return &Orchestrator{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (o *Orchestrator) Go(parent context.Context, goFunc GoFunc) {
	o.wg.Go(func() {
		uid := uuid.NewString()

		logger := logging.Get(parent).With(slog.String("goroutine-id", uid))

		logger.Debug("goroutine started")
		defer logger.Debug("goroutine ended")

		dctx := logging.Set(o.ctx, logger)

		goFunc(dctx)
	})
}

func (o *Orchestrator) Shutdown() {
	o.cancel()
	o.wg.Wait()
}
