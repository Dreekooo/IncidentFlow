package logger

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	// Initialize default logger
	// Use text handler for development, JSON for production
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Get returns the global logger instance
func Get() *slog.Logger {
	return defaultLogger
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}
