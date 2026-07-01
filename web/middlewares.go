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
	"github.com/google/uuid"
)

const RequestIdHeader = "X-Request-ID"

type requestIdCtxKey struct{}

func GetRequestId(ctx context.Context) string {
	return ctx.Value(requestIdCtxKey{}).(string)
}

func RequestId(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get(RequestIdHeader)
		if requestId == "" {
			requestId = uuid.NewString()
		}
		w.Header().Set(RequestIdHeader, requestId)

		rctx := context.WithValue(r.Context(), requestIdCtxKey{}, requestId)
		logger := logging.Get(rctx).With(slog.String("correlation-id", requestId))
		rctx = logging.Set(rctx, logger)

		next(w, r.WithContext(rctx))
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
