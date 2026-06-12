package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/AdventurerAmer/shortner/errs"
)

type Handler = func(r *http.Request) (any, int, error)
type Middleware = func(next http.HandlerFunc) http.HandlerFunc

type Mux struct {
	serveMux    *http.ServeMux
	middlewares []Middleware
}

func NewMux() *Mux {
	return &Mux{
		serveMux: &http.ServeMux{},
	}
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serve := mux.serveMux.ServeHTTP
	for _, m := range mux.middlewares {
		serve = m(serve)
	}
	serve(w, r)
}

func (mux *Mux) Post(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("POST %s", route), composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Get(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("GET %s", route), composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Put(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("PUT %s", route), composeHTTPHandlerFunc(handler))
}

func (mux *Mux) Delete(route string, handler Handler) {
	mux.serveMux.HandleFunc(fmt.Sprintf("DELETE %s", route), composeHTTPHandlerFunc(handler))
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

func composeHTTPHandlerFunc(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		var (
			resp   any
			status int
			err    error
		)
		resp, status, err = handler(r)
		if err != nil {
			var appErr *errs.Error
			if !errors.As(err, &appErr) {
				appErr = errs.Wrap(err, errs.CodeInternal, "internal server error")
			}
			status = statusfromErrCode(appErr.Code)
			resp = appErr
		}

		// TODO: ignoring errors here
		b, err := json.Marshal(resp)
		w.WriteHeader(status)
		_, err = w.Write(b)
	}
}
