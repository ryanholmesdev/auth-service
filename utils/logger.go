package utils

import (
	"context"

	"github.com/monzo/slog"
)

// InitLogger initializes the structured logger
func InitLogger() {
}

// LogInfo logs an info message with structured data
func LogInfo(ctx context.Context, msg string, data map[string]interface{}) {
	slog.Info(ctx, msg, data)
}

// LogError logs an error message with structured data
func LogError(ctx context.Context, msg string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["error"] = err
	slog.Error(ctx, msg, err, data)
}

// LogDebug logs a debug message with structured data
func LogDebug(ctx context.Context, msg string, data map[string]interface{}) {
	slog.Debug(ctx, msg, data)
}

// LogWarn logs a warning message with structured data
func LogWarn(ctx context.Context, msg string, data map[string]interface{}) {
	slog.Warn(ctx, msg, data)
}
