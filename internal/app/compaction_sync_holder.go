package app

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/langoai/lango/internal/adk"
)

// compactionSyncHolder is a swappable CompactionSyncWaiter used to bridge
// the agent build phase (where ContextAwareModelAdapter is constructed) and
// the post-wiring phase (where CompactionBuffer is constructed). It is
// wired into the adapter at build time; SetWaiter plugs the real buffer in
// once it exists.
type compactionSyncHolder struct {
	inner atomic.Pointer[adk.CompactionSyncWaiter]
}

func newCompactionSyncHolder() *compactionSyncHolder { return &compactionSyncHolder{} }

// SetWaiter sets (or replaces) the underlying waiter. Safe to call before
// or after the adapter begins receiving turns.
func (h *compactionSyncHolder) SetWaiter(w adk.CompactionSyncWaiter) {
	if w == nil {
		h.inner.Store(nil)
		return
	}
	h.inner.Store(&w)
}

// WaitForSession delegates to the current waiter. When no waiter is set
// (buffer not yet constructed, or compaction disabled), it is a no-op
// returning (true, 0) so the adapter proceeds without delay.
func (h *compactionSyncHolder) WaitForSession(ctx context.Context, key string, timeout time.Duration) (bool, time.Duration) {
	p := h.inner.Load()
	if p == nil || *p == nil {
		return true, 0
	}
	return (*p).WaitForSession(ctx, key, timeout)
}
