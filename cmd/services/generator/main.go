package main

import (
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}
	slog.Info("config was loaded successfully", "config", cfg)
}
