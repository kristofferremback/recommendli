package slogutil

import (
	"log/slog"
	"os"
	"strings"
)

func InitDefaultLogger(logLevel string) {
	level, ok := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}[strings.ToLower(logLevel)]
	if !ok {
		level = slog.LevelInfo
		defer func() { slog.Warn("Could not find log level, defaulting to info", slog.String("level", logLevel)) }()
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})))
}
