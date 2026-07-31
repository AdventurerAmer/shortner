package worker

import "github.com/AdventurerAmer/shortner/logging"

type App struct {
	logger *logging.Logger
}

func New(logger *logging.Logger) *App {
	return &App{
		logger: logger,
	}
}
