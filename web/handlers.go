package web

func (app *App) DefaultHealthHandler(c *Context) (any, error) {
	resp := struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}{
		Name:   app.Cfg.Name,
		Status: "healthy",
	}
	return resp, nil
}
