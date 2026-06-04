package logging

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AdventurerAmer/shortner/config"
)

func New(cfg *config.Config) {
	level := parseLevel(cfg.Logging.Level)
	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   *cfg.Logging.AddSource,
		ReplaceAttr: replaceAttrNonLocal,
	}
	if cfg.App.Environment == config.EnvLocal {
		opts.ReplaceAttr = replaceAttrLocal
	}
	var handler slog.Handler
	if cfg.Logging.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
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
			filename := filepath.Base(source.File)
			parentPath := filepath.Base(filepath.Dir(source.File))
			source.File = filepath.Join(parentPath, filename)
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
