package slogutil

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func Fatal(message string, args ...any) {
	logFatal(context.Background(), message, args...)
}

func FatalContext(ctx context.Context, message string, args ...any) {
	logFatal(ctx, message, args...)
}

func logFatal(ctx context.Context, message string, args ...any) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	record := slog.NewRecord(time.Now(), slog.LevelError, message, pcs[0])
	record.Add(args...)
	_ = slog.Default().Handler().Handle(ctx, record)

	os.Exit(1)
}
