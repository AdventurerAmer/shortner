package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TrackerConfig struct {
	ServiceName    string
	ServiceVersion string
	Env            config.Env
	OTLPEndpoint   string
	SampleRatio    float64
}

func NewTracker(ctx context.Context, cfg TrackerConfig) (*sdktrace.TracerProvider, error) {
	var envKeyVal attribute.KeyValue
	switch cfg.Env {
	case config.EnvLocal:
		envKeyVal = semconv.DeploymentEnvironmentNameDevelopment
	case config.EnvStaging:
		envKeyVal = semconv.DeploymentEnvironmentNameStaging
	case config.EnvProd:
		envKeyVal = semconv.DeploymentEnvironmentNameProduction
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			envKeyVal,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("'resource.Merge' failed: %w", err)
	}

	// OTLP gRPC exporter
	conn, err := grpc.NewClient(
		cfg.OTLPEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // use TLS in real prod
	)
	if err != nil {
		return nil, fmt.Errorf("'grpc.NewClient' failed: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("'otlptracegrpc.New' failed: %w", err)
	}

	// Sampler
	var sampler sdktrace.Sampler
	if cfg.SampleRatio >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRatio <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		// ParentBased + TraceIDRatioBased is the production standard
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			// Production batch settings
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxQueueSize(2048),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Propagators (W3C is standard, B3 is still common)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func NewHttpHandler(handler http.Handler, serviceName string) http.Handler {
	return otelhttp.NewHandler(handler, serviceName)
}

type Span = trace.Span
type Attribute = attribute.KeyValue

var StrAttr = attribute.String
var Int64Attr = attribute.Int64
var BoolAttr = attribute.Bool

type serviceNameCtxKey struct{}

func SetServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, serviceNameCtxKey{}, name)
}

func GetServiceName(ctx context.Context) string {
	return ctx.Value(serviceNameCtxKey{}).(string)
}

func NewSpan(ctx context.Context, name string, attrs ...Attribute) (context.Context, Span) {
	serviceName := GetServiceName(ctx)
	tr := otel.GetTracerProvider().Tracer(serviceName)
	dctx, sp := tr.Start(ctx, name, trace.WithAttributes(attrs...))
	return dctx, sp
}

func GetSpan(ctx context.Context) Span {
	return trace.SpanFromContext(ctx)
}

func GetTraceId(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() && spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func GetTracingContext(r *http.Request) context.Context {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	return ctx
}
