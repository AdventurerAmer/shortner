package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/AdventurerAmer/shortner/errs"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

func (c *Context) BindJSON(v any) error {
	contentType := c.Request.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("unsupported media type: %q expected 'application/json'", contentType)
		return errs.New(errs.CodeUnsupportedFormat, msg)
	}

	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var (
			synatxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
			maxBytesErr      *http.MaxBytesError
		)

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errs.New(errs.CodeValidation, "body contains malformed JSON")
		case errors.Is(err, io.EOF):
			return errs.New(errs.CodeValidation, "body is empty")
		case errors.As(err, &synatxErr):
			msg := fmt.Sprintf("body contains malformed JSON at character %d", synatxErr.Offset)
			return errs.New(errs.CodeValidation, msg)
		case errors.As(err, &maxBytesErr):
			return errs.New(errs.CodeValidation, "body is too large")
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				msg := fmt.Sprintf("body contains incorrect JSON type for field %q", unmarshalTypeErr.Field)
				return errs.New(errs.CodeValidation, msg)
			} else {
				msg := fmt.Sprintf("body contains malformed JSON at character %d", unmarshalTypeErr.Offset)
				return errs.New(errs.CodeValidation, msg)
			}
		}

		return fmt.Errorf("'dec.Decode' failed: %w", err)
	}

	if dec.More() {
		errs.New(errs.CodeValidation, "request body must contain only one JSON object")
	}

	return nil
}

func (c *Context) Ctx() context.Context {
	return c.Request.Context()
}

func (c *Context) SetStatus(status int) {
	c.ResponseWriter.WriteHeader(status)
}
