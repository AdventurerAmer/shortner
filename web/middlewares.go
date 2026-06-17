package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/google/uuid"
)

const RequestIdHeader = "X-Request-ID"

type requestIdCtxKey struct{}

func (app *App) RequestId(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get(RequestIdHeader)
		if requestId == "" {
			requestId = uuid.NewString()
		}
		w.Header().Set(RequestIdHeader, requestId)

		ctx := context.WithValue(r.Context(), requestIdCtxKey{}, requestId)
		next(w, r.WithContext(ctx))
	}
}

func GetRequestId(ctx context.Context) string {
	return ctx.Value(requestIdCtxKey{}).(string)
}

func (app *App) Recover(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				// TODO: pretty print stack here...
				err := fmt.Errorf("%+v", r)
				app.Logger.Error("recovering from panic", "error", err)

				resp := errs.NewInternal(err)
				status := errs.HTTPStatus(resp.Code)

				w.WriteHeader(status)

				if err := writeJSON(resp, w); err != nil {
					app.Logger.Error("failed to write resposne to client", "error", err)
				}
			}
		}()
		next(w, r)
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

func (app *App) Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next(wrappedWriter, r)

		latency := fmt.Sprintf("%d ms", time.Since(start).Milliseconds())
		app.Logger.Info(
			"HTTP Request Processed",
			"request-id", GetRequestId(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.statusCode,
			"latency", latency,
		)
	}
}
