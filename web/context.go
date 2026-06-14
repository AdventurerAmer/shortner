package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

func (c *Context) Bind(v any) error {
	if err := json.NewDecoder(c.Request.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to json decode request: %w", err)
	}
	return nil
}

func (c *Context) Ctx() context.Context {
	return c.Request.Context()
}

func (c *Context) SetStatus(status int) {
	c.ResponseWriter.WriteHeader(status)
}
