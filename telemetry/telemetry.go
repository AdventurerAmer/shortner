package telemetry

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Shutdown = func(ctx context.Context)

func New(cfg *config.Config, Name, Version string) (Shutdown, error) {
	ctx := context.Background()

	var tp *sdktrace.TracerProvider

	if cfg.Observability.Tracing.Enabled {
		var err error
		tp, err = NewTracker(ctx, TrackerConfig{
			ServiceName:    Name,
			ServiceVersion: Version,
			Env:            cfg.Env,
			OTLPEndpoint:   cfg.Observability.Tracing.Endpoint,
			SampleRatio:    cfg.Observability.Tracing.SampleRate,
		})
		if err != nil {
			return nil, fmt.Errorf("'NewTracker' failed: %w", err)
		}
		otel.SetTracerProvider(tp)
	}

	return func(ctx context.Context) {
		if tp != nil {
			if err := tp.Shutdown(ctx); err != nil {
				logger := logging.Get(ctx)
				logger.Error("'tp.Shutdown' failed", "error", err)
			}
		}
	}, nil
}
