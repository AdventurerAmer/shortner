package goorch

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/AdventurerAmer/shortner/logging"
	"github.com/google/uuid"
)

type Task = func(ctx context.Context)

type Orchestrator struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(parent context.Context) *Orchestrator {
	ctx, cancel := context.WithCancel(parent)
	return &Orchestrator{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (orch *Orchestrator) Go(parent context.Context, task Task) {
	uid := uuid.NewString()
	logger := logging.Get(parent).With(slog.String("goroutineId", uid))

	orch.wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%+v", r)
				logger.Error("recovered from panic", "err", err)
			}
		}()

		logger.Debug("goroutine started")
		defer logger.Debug("goroutine ended")

		dctx := logging.Set(orch.ctx, logger)
		task(dctx)
	})
}

func (orch *Orchestrator) Cancel() {
	orch.cancel()
}

func (orch *Orchestrator) Wait() {
	orch.wg.Wait()
}

func (orch *Orchestrator) CancelAndWait() {
	orch.cancel()
	orch.wg.Wait()
}
