package app

import (
	"context"
	"sync"
	"time"
)

// ExtendableDeadline wraps a context with a deadline that can be extended
// when agent activity is detected, up to a maximum absolute timeout.
type ExtendableDeadline struct {
	baseTimeout time.Duration
	maxTimeout  time.Duration
	start       time.Time

	mu     sync.Mutex
	timer  *time.Timer
	cancel context.CancelFunc
}

// NewExtendableDeadline creates a new ExtendableDeadline.
// baseTimeout is the initial (and per-extension) timeout duration.
// maxTimeout is the absolute maximum duration from creation time.
func NewExtendableDeadline(parent context.Context, baseTimeout, maxTimeout time.Duration) (context.Context, *ExtendableDeadline) {
	ctx, cancel := context.WithCancel(parent)

	ed := &ExtendableDeadline{
		baseTimeout: baseTimeout,
		maxTimeout:  maxTimeout,
		start:       time.Now(),
		cancel:      cancel,
	}

	// Set up initial deadline timer.
	ed.timer = time.AfterFunc(baseTimeout, func() {
		cancel()
	})

	// Also ensure we don't exceed maxTimeout from start.
	time.AfterFunc(maxTimeout, func() {
		cancel()
	})

	return ctx, ed
}

// Extend resets the deadline timer by baseTimeout from now,
// but never beyond maxTimeout from the original start time.
func (ed *ExtendableDeadline) Extend() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	elapsed := time.Since(ed.start)
	remaining := ed.maxTimeout - elapsed
	if remaining <= 0 {
		return
	}

	extension := ed.baseTimeout
	if extension > remaining {
		extension = remaining
	}

	ed.timer.Reset(extension)
}

// Stop releases the deadline resources. Must be called when done (typically via defer).
func (ed *ExtendableDeadline) Stop() {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	ed.timer.Stop()
	ed.cancel()
}
