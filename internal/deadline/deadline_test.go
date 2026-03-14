package deadline

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ExpiresOnIdle(t *testing.T) {
	t.Parallel()

	ctx, ed := New(context.Background(), 100*time.Millisecond, 5*time.Second)
	defer ed.Stop()

	select {
	case <-ctx.Done():
		// Expected: idle timeout fired
	case <-time.After(1 * time.Second):
		t.Fatal("expected context to expire on idle")
	}

	assert.Equal(t, ReasonIdle, ed.Reason())
}

func TestNew_ExpiresOnMaxTimeout(t *testing.T) {
	t.Parallel()

	// idle is long, max is short — max should fire first
	ctx, ed := New(context.Background(), 5*time.Second, 150*time.Millisecond)
	defer ed.Stop()

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("expected context to expire on max timeout")
	}

	assert.Equal(t, ReasonMaxTimeout, ed.Reason())
}

func TestExtend_ResetsIdleTimer(t *testing.T) {
	t.Parallel()

	ctx, ed := New(context.Background(), 100*time.Millisecond, 1*time.Second)
	defer ed.Stop()

	// Extend before idle timeout expires
	time.Sleep(50 * time.Millisecond)
	ed.Extend()

	// Context should still be alive after original 100ms
	time.Sleep(70 * time.Millisecond)
	require.NoError(t, ctx.Err(), "context should still be active after extension")

	// Wait for extended idle deadline to expire
	select {
	case <-ctx.Done():
		assert.Equal(t, ReasonIdle, ed.Reason())
	case <-time.After(1 * time.Second):
		t.Fatal("expected context to expire after extended deadline")
	}
}

func TestExtend_RespectsMaxTimeout(t *testing.T) {
	t.Parallel()

	maxTimeout := 200 * time.Millisecond
	ctx, ed := New(context.Background(), 100*time.Millisecond, maxTimeout)
	defer ed.Stop()

	start := time.Now()

	// Keep extending — should not exceed maxTimeout
	for i := 0; i < 10; i++ {
		time.Sleep(30 * time.Millisecond)
		ed.Extend()
	}

	<-ctx.Done()
	elapsed := time.Since(start)

	assert.Less(t, elapsed, maxTimeout+200*time.Millisecond, "should not exceed max timeout")
	assert.Equal(t, ReasonMaxTimeout, ed.Reason())
}

func TestStop_CancelsContext(t *testing.T) {
	t.Parallel()

	ctx, ed := New(context.Background(), 5*time.Second, 10*time.Second)

	ed.Stop()
	assert.Error(t, ctx.Err(), "context should be canceled after Stop")
	assert.Equal(t, ReasonCancelled, ed.Reason())
}

func TestExtend_AfterDone_IsNoop(t *testing.T) {
	t.Parallel()

	_, ed := New(context.Background(), 50*time.Millisecond, 5*time.Second)
	defer ed.Stop()

	// Wait for idle expiry
	time.Sleep(100 * time.Millisecond)

	// Should not panic
	ed.Extend()
	assert.Equal(t, ReasonIdle, ed.Reason())
}
