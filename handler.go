package fingerscrossed

import (
	"context"
	"errors"
	"log/slog"
)

type Handler interface {
	slog.Handler
	FlushLogs(minimumLevel slog.Level) error
}

func NewHandler(wrapping slog.Handler, options ...option) Handler {
	h := &handler{
		thresholdLevel: slog.LevelError,
		wrapped:        wrapping,
		stored:         &recordBuffer{},
	}
	for _, option := range options {
		option(h)
	}
	return h
}

type option func(*handler)

// WithThresholdLevel modifies the level at which the handler will flush its logs and stop buffering them.
// Default is slog.LevelError
func WithThresholdLevel(level slog.Level) option {
	return func(h *handler) {
		h.thresholdLevel = level
	}
}

type handler struct {
	thresholdLevel slog.Level
	wrapped        slog.Handler
	stored         *recordBuffer
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= h.thresholdLevel {
		return errors.Join(
			h.stored.Unbuffer(),
			h.wrapped.Handle(ctx, record),
		)
	}
	return h.stored.Handle(ctx, h.wrapped, record)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{
		thresholdLevel: h.thresholdLevel,
		wrapped:        h.wrapped.WithAttrs(attrs),
		stored:         h.stored,
	}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{
		thresholdLevel: h.thresholdLevel,
		wrapped:        h.wrapped.WithGroup(name),
		stored:         h.stored,
	}
}

func (h *handler) FlushLogs(minimumLevel slog.Level) error {
	return h.stored.FlushLogs(minimumLevel)
}
