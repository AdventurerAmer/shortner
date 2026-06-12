package web

import (
	"context"
	"net/http"

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
