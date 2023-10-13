package fingerscrossed

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
)

type record struct {
	fn    func() error
	level slog.Level
}

type recordBuffer struct {
	lock       sync.Mutex
	records    []record
	unbuffered atomic.Bool
}

func (b *recordBuffer) Handle(ctx context.Context, h slog.Handler, r slog.Record) error {
	if b.unbuffered.Load() {
		return h.Handle(ctx, r)
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.records = append(b.records, record{
		fn: func() error {
			return h.Handle(ctx, r)
		},
		level: r.Level,
	})

	return nil
}

func (b *recordBuffer) Unbuffer() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.unbuffered.Store(true)

	errs := make([]error, 0, len(b.records))
	for _, r := range b.records {
		errs = append(errs, r.fn())
	}
	b.records = nil
	return errors.Join(errs...)
}

func (b *recordBuffer) FlushLogs(level slog.Level) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	errs := make([]error, 0, len(b.records))
	for _, r := range b.records {
		if r.level >= level {
			errs = append(errs, r.fn())
		}
	}
	b.records = nil
	return errors.Join(errs...)
}
