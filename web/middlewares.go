package web

import (
	"context"
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
			if err := recover(); err != nil {
				app.Logger.Error("recovering from panic", "error", err)

				resp := errs.New(errs.CodeInternal, "internal server error")
				status := statusfromErrCode(resp.Code)
				if err := writeJSON(resp, status, w); err != nil {
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

		latency := time.Since(start)
		app.Logger.Info("HTTP Request Processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.statusCode,
			"latency", latency,
		)
	}
}
