package logger

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init(serviceName string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	// Use JSON handler for production-style logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler).With("service", serviceName)
	slog.SetDefault(Log)
}

func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

func Fatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}

func WithCtx(ctx context.Context) *slog.Logger {
	// In a real implementation, extract trace_id from context
	return Log
}
