package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/logging"
)

type Handler = func(c *Context) (any, error)
type Middleware = func(next http.HandlerFunc) http.HandlerFunc

type Mux struct {
	logger      *logging.Logger
	serveMux    *http.ServeMux
	middlewares []Middleware
}

func NewMux(logger *logging.Logger) *Mux {
	return &Mux{
		logger:   logger,
		serveMux: &http.ServeMux{},
	}
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serve := mux.serveMux.ServeHTTP
	for _, m := range slices.Backward(mux.middlewares) {
		serve = m(serve)
	}
	serve(w, r)
}

func (mux *Mux) Post(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("POST %s", route), mux.composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Get(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("GET %s", route), mux.composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Put(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("PUT %s", route), mux.composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Delete(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("DELETE %s", route), mux.composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Use(m Middleware) {
	mux.middlewares = append(mux.middlewares, m)
}

func statusfromErrCode(code errs.Code) int {
	switch code {
	case errs.CodeInternal:
		return http.StatusInternalServerError
	case errs.CodeResourceNotFound:
		return http.StatusNotFound
	}
	return 0
}

func (mux *Mux) composeHTTPHandlerFunc(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := &Context{
			Request:        r,
			ResponseWriter: w,
		}
		resp, err := handler(c)
		if err != nil {
			var appErr *errs.Error
			if !errors.As(err, &appErr) {
				appErr = errs.Wrap(err, errs.CodeInternal, "internal server error")
				mux.logger.Error("internal server error", "error", err)
			}
			status := statusfromErrCode(appErr.Code)
			w.WriteHeader(status)
			resp = appErr
		}
		if err := writeJSON(resp, w); err != nil {
			mux.logger.Error("failed to write resposne to client", "error", err)
		}
	}
}

func writeJSON(resp any, w http.ResponseWriter) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshaling payload to json failed: %w", err)
	}

	w.Header().Add("Content-Type", "application/json")

	b := &bytes.Buffer{}
	if _, err := b.Write(payload); err != nil {
		return fmt.Errorf("writing marshaled payload to a buffer failed: %w", err)
	}

	if _, err = w.Write(b.Bytes()); err != nil {
		return fmt.Errorf("writing response failed: %w", err)
	}

	return nil
}
