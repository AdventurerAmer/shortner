package telemetry

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	instruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type MetricsConfig struct {
	Endpoint string
}

var (
	AllocatedMemory metric.Float64ObservableGauge
	NumGorouties    metric.Int64ObservableGauge
	RequestsLatency metric.Int64Histogram
	RequestsCounter metric.Int64Counter
)

func NewPrometheus(ctx context.Context, res *resource.Resource, cfg MetricsConfig) (*sdkmetric.MeterProvider, error) {
	// TODO: use TLS in real prod
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("'otlpmetricgrpc.New' failed: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(15*time.Second))),
	)

	if err := instruntime.Start(
		instruntime.WithMinimumReadMemStatsInterval(15*time.Second),
		instruntime.WithMeterProvider(mp),
	); err != nil {
		return nil, fmt.Errorf("'insruntime.Start' failed: %w", err)
	}

	var attributes = []attribute.KeyValue{
		attribute.Key("application").String(serviceName),
		attribute.Key("container_id").String(os.Getenv("HOSTNAME")),
	}

	AllocatedMemory, err = NewFloat64Guage("allocated_memory", "Amount of memory used.", "MB")
	if err != nil {
		return nil, fmt.Errorf("failed to create allocated memory guage: %w", err)
	}

	meter := GetMeter()

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			allocatedMemoryMB := float64(stats.Sys) / 1_048_576 // Convert bytes to MB
			o.ObserveFloat64(AllocatedMemory, allocatedMemoryMB, metric.WithAttributes(attributes...))
			return nil
		},
		AllocatedMemory,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register allocated memory guage callback: %w", err)
	}

	NumGorouties, err = NewInt64Guage("num_gorouties", "Number of running goruties.", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create num gorouties guage: %w", err)
	}

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(NumGorouties, int64(runtime.NumGoroutine()), metric.WithAttributes(attributes...))
			return nil
		},
		NumGorouties,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create num gorouties guage: %w", err)
	}

	RequestsLatency, err = NewInt64Hisogram("requests_latency", "The duration of requests", "ms")
	if err != nil {
		return nil, fmt.Errorf("failed to create requests histogram: %w", err)
	}

	RequestsCounter, err = NewInt64Counter("requests_total", "Total number of requests.")
	if err != nil {
		return nil, fmt.Errorf("failed to create requests counter: %w", err)
	}

	return mp, nil
}

func GetMeter() metric.Meter {
	return otel.GetMeterProvider().Meter(serviceName)
}

type Int64Counter = metric.Int64Counter
type Float64Counter = metric.Float64Counter

func NewInt64Counter(name, description string) (Int64Counter, error) {
	meter := GetMeter()
	counter, err := meter.Int64Counter(fmt.Sprintf("%s_%s", serviceName, name), metric.WithDescription(description))
	if err != nil {
		return nil, err
	}
	return counter, nil
}

func NewFloat64Counter(name, description string) (Float64Counter, error) {
	meter := GetMeter()
	counter, err := meter.Float64Counter(fmt.Sprintf("%s_%s", serviceName, name), metric.WithDescription(description))
	if err != nil {
		return nil, err
	}
	return counter, nil
}

type Int64Histogram = metric.Int64Histogram
type Float64Histogram = metric.Float64Histogram

func NewInt64Hisogram(name, description, unit string) (Int64Histogram, error) {
	meter := GetMeter()
	histogram, err := meter.Int64Histogram(fmt.Sprintf("%s_%s", serviceName, name), metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil {
		return nil, err
	}
	return histogram, nil
}

func NewFloat64Hisogram(name, description, unit string) (Float64Histogram, error) {
	meter := GetMeter()
	histogram, err := meter.Float64Histogram(fmt.Sprintf("%s_%s", serviceName, name), metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil {
		return nil, err
	}
	return histogram, nil
}

type Int64Guage = metric.Int64ObservableGauge
type Float64Guage = metric.Float64ObservableGauge

func NewInt64Guage(name, description, unit string) (Int64Guage, error) {
	meter := GetMeter()
	guage, err := meter.Int64ObservableGauge(fmt.Sprintf("%s_%s", serviceName, name), metric.WithUnit(unit), metric.WithDescription(description))
	if err != nil {
		return nil, err
	}
	return guage, nil
}

func NewFloat64Guage(name, description, unit string) (Float64Guage, error) {
	meter := GetMeter()
	guage, err := meter.Float64ObservableGauge(fmt.Sprintf("%s_%s", serviceName, name), metric.WithUnit(unit), metric.WithDescription(description))
	if err != nil {
		return nil, err
	}
	return guage, nil
}
