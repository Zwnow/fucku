package pkg

import (
	"context"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (m *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		r := record
		if err := h.Handle(ctx, r); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var wrapped []slog.Handler
	for _, h := range m.handlers {
		wrapped = append(wrapped, h.WithAttrs(attrs))
	}

	return &MultiHandler{handlers: wrapped}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	var wrapped []slog.Handler
	for _, h := range m.handlers {
		wrapped = append(wrapped, h.WithGroup(name))
	}
	return &MultiHandler{handlers: wrapped}
}

// Sets up a Slog logger that writes to stdout and a rotating file
func NewLogger(logPath string, level slog.Leveler) *slog.Logger {
	// Setup file logger with rotation
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // Megabytes
		MaxBackups: 10,
		MaxAge:     60,
		Compress:   true,
	}

	// Handlers
	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	fileHandler := slog.NewTextHandler(fileWriter, &slog.HandlerOptions{
		Level: level,
	})

	// Combine handlers
	multiHandler := NewMultiHandler(stdoutHandler, fileHandler)

	return slog.New(multiHandler)
}
