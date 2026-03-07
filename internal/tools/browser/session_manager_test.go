package browser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionManager_EnsureSession_CreatesOnce(t *testing.T) {
	t.Parallel()

	tool, err := New(Config{
		Headless:       true,
		SessionTimeout: 5 * time.Minute,
	})
	require.NoError(t, err)

	sm := NewSessionManager(tool)
	defer sm.Close()

	// First call creates a session
	id1, err := sm.EnsureSession()
	if err != nil {
		// Browser may not be available in CI; skip gracefully
		t.Skipf("browser not available: %v", err)
	}
	require.NotEmpty(t, id1)

	// Second call reuses the same session
	id2, err := sm.EnsureSession()
	require.NoError(t, err)
	assert.Equal(t, id1, id2)
}

func TestSessionManager_Close(t *testing.T) {
	t.Parallel()

	tool, err := New(Config{
		Headless:       true,
		SessionTimeout: 5 * time.Minute,
	})
	require.NoError(t, err)

	sm := NewSessionManager(tool)

	// Close without any session should not error
	require.NoError(t, sm.Close())
}

func TestSessionManager_Tool(t *testing.T) {
	t.Parallel()

	tool, err := New(Config{
		Headless:       true,
		SessionTimeout: 5 * time.Minute,
	})
	require.NoError(t, err)

	sm := NewSessionManager(tool)
	assert.Equal(t, tool, sm.Tool())
}
