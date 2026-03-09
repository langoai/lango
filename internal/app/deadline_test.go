package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtendableDeadline_ExpiresWithoutExtension(t *testing.T) {
	t.Parallel()

	ctx, ed := NewExtendableDeadline(context.Background(), 100*time.Millisecond, 500*time.Millisecond)
	defer ed.Stop()

	select {
	case <-ctx.Done():
		// Expected: deadline expired after ~100ms
	case <-time.After(1 * time.Second):
		t.Fatal("expected context to expire")
	}

	assert.Error(t, ctx.Err())
}

func TestExtendableDeadline_ExtendsProperly(t *testing.T) {
	t.Parallel()

	ctx, ed := NewExtendableDeadline(context.Background(), 100*time.Millisecond, 1*time.Second)
	defer ed.Stop()

	// Extend before the 100ms base timeout expires.
	time.Sleep(50 * time.Millisecond)
	ed.Extend()

	// The context should still be alive after original 100ms.
	time.Sleep(70 * time.Millisecond)
	assert.NoError(t, ctx.Err(), "context should still be active after extension")

	// Wait for extended deadline to expire.
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("expected context to expire after extended deadline")
	}
}

func TestExtendableDeadline_RespectsMaxTimeout(t *testing.T) {
	t.Parallel()

	maxTimeout := 200 * time.Millisecond
	ctx, ed := NewExtendableDeadline(context.Background(), 100*time.Millisecond, maxTimeout)
	defer ed.Stop()

	start := time.Now()

	// Keep extending — should not exceed maxTimeout.
	for i := 0; i < 10; i++ {
		time.Sleep(30 * time.Millisecond)
		ed.Extend()
	}

	<-ctx.Done()
	elapsed := time.Since(start)

	// Should not exceed maxTimeout + generous tolerance for CI scheduling jitter.
	assert.Less(t, elapsed, maxTimeout+200*time.Millisecond, "should not exceed max timeout")
}

func TestExtendableDeadline_StopCancelsContext(t *testing.T) {
	t.Parallel()

	ctx, ed := NewExtendableDeadline(context.Background(), 5*time.Second, 10*time.Second)

	ed.Stop()
	assert.Error(t, ctx.Err(), "context should be canceled after Stop")
}
