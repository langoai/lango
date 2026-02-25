package types

import (
	"context"
	"time"
)

// detachedCtx wraps context.Background() but delegates Value() to the
// original parent. This preserves session keys, approval targets, and
// other request-scoped values while decoupling the lifetime from the
// parent context (Done/Err/Deadline all follow context.Background()).
type detachedCtx struct {
	parent context.Context
}

func (c *detachedCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *detachedCtx) Done() <-chan struct{}        { return nil }
func (c *detachedCtx) Err() error                   { return nil }
func (c *detachedCtx) Value(key any) any            { return c.parent.Value(key) }

// DetachContext returns a new context that is independent of the parent's
// cancellation and deadline but preserves all context values.
// Use this when spawning long-running goroutines that must not be
// cancelled when the originating request completes.
func DetachContext(ctx context.Context) context.Context {
	return &detachedCtx{parent: ctx}
}
