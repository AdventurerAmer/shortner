package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

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
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(15*time.Second)),
		),
	)

	return mp, nil
}

type Observer metric.Observer
type Float64Observable = metric.Float64Observable
type Int64Observable = metric.Int64Observable

type ObserveCallback func(ctx context.Context, observer Observer) error

func ObserveFloat64(observer Observer, observable Float64Observable, value float64) {
	attributes := []attribute.KeyValue{
		attribute.Key("application").String(serviceName),
		attribute.Key("container_id").String(os.Getenv("HOSTNAME")),
	}
	observer.ObserveFloat64(observable, value, metric.WithAttributes(attributes...))
}

func ObserveInt64(observer Observer, observable Int64Observable, value int64) {
	attributes := []attribute.KeyValue{
		attribute.Key("application").String(serviceName),
		attribute.Key("container_id").String(os.Getenv("HOSTNAME")),
	}
	observer.ObserveInt64(observable, value, metric.WithAttributes(attributes...))
}

func Observe(callback ObserveCallback, observables ...metric.Observable) error {
	meter := GetMeter()
	if _, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			return callback(ctx, o)
		},
		observables...,
	); err != nil {
		return fmt.Errorf("failed to register allocated memory guage callback: %w", err)
	}
	return nil
}

func GetMeter() metric.Meter {
	return otel.GetMeterProvider().Meter(serviceName)
}

type Int64Counter = metric.Int64Counter
type Float64Counter = metric.Float64Counter

func NewInt64Counter(name, description string) (Int64Counter, error) {
	meter := GetMeter()
	counter, err := meter.Int64Counter(
		fmt.Sprintf("%s_%s", serviceName, name),
		metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("'meter.Int64Counter' failed: %w", err)
	}
	return counter, nil
}

func NewFloat64Counter(name, description string) (Float64Counter, error) {
	meter := GetMeter()
	counter, err := meter.Float64Counter(
		fmt.Sprintf("%s_%s", serviceName, name),
		metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("'meter.Float64Counter' failed: %w", err)
	}
	return counter, nil
}

type Int64Histogram = metric.Int64Histogram
type Float64Histogram = metric.Float64Histogram

func NewInt64Hisogram(name, description, unit string) (Int64Histogram, error) {
	meter := GetMeter()
	histogram, err := meter.Int64Histogram(
		fmt.Sprintf("%s_%s", serviceName, name),
		metric.WithDescription(description),
		metric.WithUnit(unit))
	if err != nil {
		return nil, fmt.Errorf("'meter.Int64Histogram' failed: %w", err)
	}
	return histogram, nil
}

func NewFloat64Hisogram(name, description, unit string) (Float64Histogram, error) {
	meter := GetMeter()
	histogram, err := meter.Float64Histogram(
		fmt.Sprintf("%s_%s", serviceName, name),
		metric.WithDescription(description),
		metric.WithUnit(unit))
	if err != nil {
		return nil, fmt.Errorf("'meter.Float64Histogram' failed: %w", err)
	}
	return histogram, nil
}

type Int64Guage = metric.Int64ObservableGauge
type Float64Guage = metric.Float64ObservableGauge

func NewInt64Guage(name, description, unit string) (Int64Guage, error) {
	meter := GetMeter()
	guage, err := meter.Int64ObservableGauge(
		fmt.Sprintf("%s_%s", serviceName, name),
		metric.WithUnit(unit),
		metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("'meter.Int64ObservableGauge' failed: %w", err)
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
