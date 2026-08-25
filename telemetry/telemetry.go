package telemetry

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

var serviceName string

type Shutdown = func(ctx context.Context)

func New(cfg *config.Config, Name, Version string) (Shutdown, error) {
	serviceName = Name

	ctx := context.Background()

	var envAtrrib attribute.KeyValue
	switch cfg.Env {
	case config.EnvLocal:
		envAtrrib = semconv.DeploymentEnvironmentNameDevelopment
	case config.EnvStaging:
		envAtrrib = semconv.DeploymentEnvironmentNameStaging
	case config.EnvProd:
		envAtrrib = semconv.DeploymentEnvironmentNameProduction
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(Name),
			semconv.ServiceVersion(Version),
			envAtrrib,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("'resource.Merge' failed: %w", err)
	}

	var tp *sdktrace.TracerProvider

	if cfg.Observability.Tracing.Enabled {
		trackerCfg := TrackerConfig{
			OTLPEndpoint: cfg.Observability.Tracing.Endpoint,
			SampleRatio:  cfg.Observability.Tracing.SampleRate,
		}
		var err error
		tp, err = NewTracker(ctx, res, trackerCfg)
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

	var mp *metric.MeterProvider

	if cfg.Observability.Metrics.Enabled {
		metricsCfg := MetricsConfig{
			Endpoint: cfg.Observability.Metrics.Endpoint,
		}
		var err error
		mp, err = NewPrometheus(ctx, res, metricsCfg)
		if err != nil {
			return nil, fmt.Errorf("'NewPrometheus' failed: %w", err)
		}
		otel.SetMeterProvider(mp)
	}

	return func(ctx context.Context) {
		logger := logging.Get(ctx)

		if tp != nil {
			if err := tp.Shutdown(ctx); err != nil {
				logger.Error("'tp.Shutdown' failed", "error", err)
			}
		}
		if mp != nil {
			if err := mp.Shutdown(ctx); err != nil {
				logger.Error("'mp.Shutdown' failed", "error", err)
			}
		}
	}, nil
}
