package telemetry

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var serviceName string

type Shutdown = func(ctx context.Context)

func New(cfg *config.Config, Name, Version string) (Shutdown, error) {
	serviceName = Name

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

		// Propagators (W3C is standard, B3 is still common)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
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
