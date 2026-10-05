package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var defaultLogger *slog.Logger

var mu sync.RWMutex

func newLogger(w io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler)
}

func init() {
	mu.Lock()
	defer mu.Unlock()

	defaultLogger = newLogger(os.Stdout)
	slog.SetDefault(defaultLogger)
}

// SetOutput changes the logger output destination.
// Useful in tests to capture or silence logs.
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	defaultLogger = newLogger(w)
	slog.SetDefault(defaultLogger)
}

// Get returns the global logger instance
func Get() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()

	return defaultLogger
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()

	defaultLogger.Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()

	defaultLogger.Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()

	defaultLogger.Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()

	defaultLogger.Error(msg, args...)
}
