package main

import (
	"fmt"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	logging.New(cfg)

	logging.Info("config was loaded successfully", "environment", cfg.App.Environment, "version", cfg.App.Version)
}
