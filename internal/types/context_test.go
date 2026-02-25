package types

import (
	"context"
	"testing"
	"time"
)

type testKey struct{}

func TestDetachContext_ParentCancelDoesNotAffectChild(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	detached := DetachContext(parent)

	cancel() // cancel parent

	if parent.Err() == nil {
		t.Fatal("parent should be cancelled")
	}
	if detached.Err() != nil {
		t.Fatalf("detached context should not be cancelled, got: %v", detached.Err())
	}
	if detached.Done() != nil {
		t.Fatal("detached Done channel should be nil")
	}
}

func TestDetachContext_PreservesValues(t *testing.T) {
	parent := context.WithValue(context.Background(), testKey{}, "hello")
	detached := DetachContext(parent)

	got := detached.Value(testKey{})
	if got != "hello" {
		t.Fatalf("expected value 'hello', got %v", got)
	}
}

func TestDetachContext_NoDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	detached := DetachContext(parent)

	if _, ok := detached.Deadline(); ok {
		t.Fatal("detached context should have no deadline")
	}
}

func TestDetachContext_WithCancelWrapping(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	detached := DetachContext(parent)
	child, childCancel := context.WithCancel(detached)

	// Cancel parent — child should be unaffected.
	parentCancel()
	if child.Err() != nil {
		t.Fatal("child of detached should not be cancelled when parent is cancelled")
	}

	// Cancel child directly — should work.
	childCancel()
	if child.Err() == nil {
		t.Fatal("child should be cancelled after childCancel()")
	}
}

func TestDetachContext_WithTimeoutWrapping(t *testing.T) {
	parent := context.WithValue(context.Background(), testKey{}, "timeout-test")
	detached := DetachContext(parent)
	child, cancel := context.WithTimeout(detached, 50*time.Millisecond)
	defer cancel()

	// Value should propagate through detached → child.
	if child.Value(testKey{}) != "timeout-test" {
		t.Fatal("value should propagate through detached context")
	}

	// Wait for timeout.
	<-child.Done()
	if child.Err() != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", child.Err())
	}
}
