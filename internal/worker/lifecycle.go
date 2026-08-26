package worker

import (
	"context"
	"sync"
	"time"
)

type Lifecycle struct {
	mu                   sync.Mutex
	started, stopped     bool
	startedAt, stoppedAt time.Time
}

func (l *Lifecycle) Start(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.started && !l.stopped {
		return false
	}
	l.started = true
	l.stopped = false
	l.startedAt = now
	l.stoppedAt = time.Time{}
	return true
}
func (l *Lifecycle) Stop(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.started || l.stopped {
		return false
	}
	l.stopped = true
	l.stoppedAt = now
	return true
}
func (l *Lifecycle) Running() bool { l.mu.Lock(); defer l.mu.Unlock(); return l.started && !l.stopped }
func RunUntilCancel(ctx context.Context, interval time.Duration, work func(context.Context)) error {
	if interval <= 0 {
		return context.Canceled
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			work(ctx)
		}
	}
}
