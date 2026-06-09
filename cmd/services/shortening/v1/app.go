package v1

import (
	"fmt"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	logger := logging.New(cfg)

	app := web.New(logger, &cfg.Services.Shortening)

	mux := web.NewMux()
	mux.Get("/health", app.DefaultHealthHandler)

	app.Run(mux)

	return 0
}
