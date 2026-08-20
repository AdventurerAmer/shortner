package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func Trace(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract remote context from request headers
		ctx := otel.GetTextMapPropagator().Extract(
			r.Context(),
			propagation.HeaderCarrier(r.Header),
		)

		// 2. Create span name (common pattern)
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		sctx, span := telemetry.NewSpan(ctx, spanName, semconv.HTTPMethodKey.String(r.Method),
			semconv.HTTPTargetKey.String(r.URL.Path),
			semconv.HTTPSchemeKey.String(r.URL.Scheme),
			semconv.NetHostNameKey.String(r.Host),
			attribute.String("http.user_agent", r.UserAgent()))

		defer span.End()

		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		traceId := telemetry.GetTraceId(sctx)
		logger := logging.Get(r.Context()).With(slog.String("correlationId", traceId))
		dctx := logging.Set(r.Context(), logger)

		next(rw, r.WithContext(dctx))

		// 6. Set final status
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(rw.statusCode))
		if rw.statusCode >= 500 {
			span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
		}
	}
}

func Recover(env config.Env) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger := logging.Get(r.Context())

					err := fmt.Errorf("%+v", rec)

					if env == config.EnvProd {
						logger.Error("recovered from panic", "error", err)
					} else {
						stackTrace := string(debug.Stack())
						logger.Error("recovered from panic", "error", err, "stack-trace", stackTrace)
					}
					resp := errs.NewInternal(err)
					status := errs.HTTPStatus(resp.Code)

					w.WriteHeader(status)

					if err := writeJSON(resp, w); err != nil {
						logger.Error("failed to write resposne to client", "error", err)
					}
				}
			}()
			next(w, r)
		}
	}
}

func Timeout(timeout time.Duration) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			dctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next(w, r.WithContext(dctx))
		}
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next(wrappedWriter, r)

		latency := fmt.Sprintf("%dms", time.Since(start).Milliseconds())
		logger := logging.Get(r.Context())
		logger.Info(
			"HTTP Request Processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status-code", wrappedWriter.statusCode,
			"status", http.StatusText(wrappedWriter.statusCode),
			"latency", latency,
		)
	}
}
