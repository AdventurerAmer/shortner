package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AdventurerAmer/shortner/config"
)

type Logger = slog.Logger

func New(cfg *config.Config) *Logger {
	level := parseLevel(cfg.Logging.Level)

	replaceAttr := replaceAttrNonLocal
	if cfg.App.Environment == config.EnvLocal {
		replaceAttr = replaceAttrLocal
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   *cfg.Logging.AddSource,
		ReplaceAttr: replaceAttr,
	}

	var handler slog.Handler
	if cfg.Logging.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func parseLevel(level string) slog.Level {
	switch level {
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func replaceAttrLocal(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.SourceKey:
		if source, ok := attr.Value.Any().(*slog.Source); ok {
			wd, err := os.Getwd()
			if err != nil {
				return attr
			}
			rel, err := filepath.Rel(wd, source.File)
			if err != nil {
				return attr
			}
			source.File = rel
		}
	case slog.TimeKey:
		if len(groups) == 0 {
			if t, ok := attr.Value.Any().(time.Time); ok {
				formattedTime := t.Format("Monday, Jan _2, 2006 at 3:04PM")
				return slog.String(slog.TimeKey, formattedTime)
			}
		}
	}
	return attr
}

func replaceAttrNonLocal(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case "password", "secret", "token", "jwt_secret":
		return slog.String(attr.Key, "[REDACTED]")
	}
	return attr
}
