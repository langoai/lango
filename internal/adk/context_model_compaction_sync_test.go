package adk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeSyncWaiter is a controllable CompactionSyncWaiter for tests.
type fakeSyncWaiter struct {
	done bool
	wait time.Duration
	seen bool
}

func (f *fakeSyncWaiter) WaitForSession(_ context.Context, _ string, _ time.Duration) (bool, time.Duration) {
	f.seen = true
	return f.done, f.wait
}

func TestCompactionSyncHolderContract(t *testing.T) {
	t.Parallel()

	f := &fakeSyncWaiter{done: true, wait: 10 * time.Millisecond}
	ok, waited := f.WaitForSession(context.Background(), "sess", time.Second)
	assert.True(t, ok)
	assert.Equal(t, 10*time.Millisecond, waited)
	assert.True(t, f.seen)
}

func TestCompactionSyncWaiterTimeoutShape(t *testing.T) {
	t.Parallel()

	f := &fakeSyncWaiter{done: false, wait: time.Second}
	ok, waited := f.WaitForSession(context.Background(), "sess", time.Second)
	assert.False(t, ok)
	assert.Equal(t, time.Second, waited)
}
