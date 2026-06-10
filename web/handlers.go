package web

import "net/http"

func (app *App) DefaultHealthHandler(r *http.Request) (any, int, error) {
	resp := struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}{
		Name:   app.Cfg.Name,
		Status: "healthy",
	}
	return resp, http.StatusOK, nil
}
