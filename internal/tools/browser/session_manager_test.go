package browser

import (
	"testing"
	"time"
)

func TestSessionManager_EnsureSession_CreatesOnce(t *testing.T) {
	tool, err := New(Config{
		Headless:       true,
		SessionTimeout: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("new tool: %v", err)
	}

	sm := NewSessionManager(tool)
	defer sm.Close()

	// First call creates a session
	id1, err := sm.EnsureSession()
	if err != nil {
		// Browser may not be available in CI; skip gracefully
		t.Skipf("browser not available: %v", err)
	}
	if id1 == "" {
		t.Fatal("expected non-empty session ID")
	}

	// Second call reuses the same session
	id2, err := sm.EnsureSession()
	if err != nil {
		t.Fatalf("ensure session (2nd): %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same session ID, got %q and %q", id1, id2)
	}
}

func TestSessionManager_Close(t *testing.T) {
	tool, err := New(Config{
		Headless:       true,
		SessionTimeout: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("new tool: %v", err)
	}

	sm := NewSessionManager(tool)

	// Close without any session should not error
	if err := sm.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestSessionManager_Tool(t *testing.T) {
	tool, err := New(Config{
		Headless:       true,
		SessionTimeout: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("new tool: %v", err)
	}

	sm := NewSessionManager(tool)
	if sm.Tool() != tool {
		t.Error("Tool() should return the underlying tool")
	}
}
