package deadline

import (
	"context"
	"sync"
	"time"
)

// Reason describes why the deadline expired.
type Reason string

const (
	ReasonIdle       Reason = "idle"        // no activity within idle timeout
	ReasonMaxTimeout Reason = "max_timeout" // hard ceiling reached
	ReasonCancelled  Reason = "cancelled"   // explicitly stopped or parent cancelled
)

// ExtendableDeadline wraps a context with a deadline that can be extended
// when agent activity is detected, up to a maximum absolute timeout.
type ExtendableDeadline struct {
	idleTimeout time.Duration
	maxTimeout  time.Duration
	start       time.Time

	mu       sync.Mutex
	parent   context.Context
	timer    *time.Timer
	maxTimer *time.Timer
	cancel   context.CancelFunc
	reason   Reason
	done     bool
	clamped  bool // true when idle extension was clamped to remaining max time
}

// New creates a new ExtendableDeadline.
// idleTimeout is the initial (and per-extension) timeout duration for inactivity.
// maxTimeout is the absolute maximum duration from creation time.
func New(parent context.Context, idleTimeout, maxTimeout time.Duration) (context.Context, *ExtendableDeadline) {
	ctx, cancel := context.WithCancel(parent)

	ed := &ExtendableDeadline{
		idleTimeout: idleTimeout,
		maxTimeout:  maxTimeout,
		start:       time.Now(),
		parent:      parent,
		cancel:      cancel,
		reason:      ReasonIdle, // default reason if idle timer fires
	}

	// Set up idle deadline timer.
	ed.timer = time.AfterFunc(idleTimeout, func() {
		ed.mu.Lock()
		defer ed.mu.Unlock()
		if !ed.done {
			if ed.clamped {
				ed.reason = ReasonMaxTimeout
			} else {
				ed.reason = ReasonIdle
			}
			ed.done = true
			cancel()
		}
	})

	// Also ensure we don't exceed maxTimeout from start.
	ed.maxTimer = time.AfterFunc(maxTimeout, func() {
		ed.mu.Lock()
		defer ed.mu.Unlock()
		if !ed.done {
			ed.reason = ReasonMaxTimeout
			ed.done = true
			cancel()
		}
	})

	go ed.watchParent(parent, ctx)
	if parent.Err() != nil {
		ed.cancelWithReason(ReasonCancelled)
	}

	return ctx, ed
}

func (ed *ExtendableDeadline) watchParent(parent context.Context, ctx context.Context) {
	select {
	case <-parent.Done():
		ed.cancelWithReason(ReasonCancelled)
	case <-ctx.Done():
		if parent.Err() != nil {
			ed.cancelWithReason(ReasonCancelled)
		}
	}
}

// Extend resets the idle timer by idleTimeout from now,
// but never beyond maxTimeout from the original start time.
func (ed *ExtendableDeadline) Extend() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if ed.done {
		return
	}

	elapsed := time.Since(ed.start)
	remaining := ed.maxTimeout - elapsed
	if remaining <= 0 {
		return
	}

	extension := ed.idleTimeout
	if extension > remaining {
		extension = remaining
		ed.clamped = true
	}

	ed.timer.Reset(extension)
}

// Stop releases the deadline resources. Must be called when done (typically via defer).
func (ed *ExtendableDeadline) Stop() {
	ed.cancelWithReason(ReasonCancelled)
}

// Reason returns the reason the deadline expired (or ReasonCancelled if Stop was called).
func (ed *ExtendableDeadline) Reason() Reason {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if !ed.done && ed.parent != nil && ed.parent.Err() != nil {
		ed.reason = ReasonCancelled
		ed.done = true
		ed.stopTimersLocked()
		ed.cancel()
	}
	return ed.reason
}

func (ed *ExtendableDeadline) cancelWithReason(reason Reason) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if !ed.done {
		ed.reason = reason
		ed.done = true
	}
	ed.stopTimersLocked()
	ed.cancel()
}

func (ed *ExtendableDeadline) stopTimersLocked() {
	if ed.timer != nil {
		ed.timer.Stop()
	}
	if ed.maxTimer != nil {
		ed.maxTimer.Stop()
	}
}
