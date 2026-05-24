package smartaccount

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCreateRejectsInvalidOutputBeforeLoading(t *testing.T) {
	original := executeSessionCreate
	called := false
	executeSessionCreate = func(_ BootLoader, _, _ []string, _, _ string) (sessionCreateResult, func(), error) {
		called = true
		return sessionCreateResult{}, nil, assert.AnError
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestSessionCreateRunsCleanupAfterSuccessfulOutput(t *testing.T) {
	original := executeSessionCreate
	cleanupCalled := false
	executeSessionCreate = func(_ BootLoader, targets, functions []string, limit, duration string) (sessionCreateResult, func(), error) {
		assert.Equal(t, []string{"0x000000000000000000000000000000000000aaaa"}, targets)
		assert.Equal(t, []string{"0xa9059cbb"}, functions)
		assert.Equal(t, "42", limit)
		assert.Equal(t, "2h", duration)
		return sessionCreateResult{
			ID:        "session-12345678",
			Address:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			Targets:   []string{"0x000000000000000000000000000000000000aaaa"},
			Functions: []string{"0xa9059cbb"},
			Limit:     "42",
			ExpiresAt: "2026-05-15T02:00:00Z",
			CreatedAt: "2026-05-15T00:00:00Z",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd,
		"--targets", "0x000000000000000000000000000000000000aaaa",
		"--functions", "0xa9059cbb",
		"--limit", "42",
		"--duration", "2h",
	)

	require.NoError(t, err)
	assert.Contains(t, out, "Session Key Created")
	assert.Contains(t, out, "Spend Limit:")
	assert.True(t, cleanupCalled)
}

func TestSessionCreatePropagatesExecutorErrorWithoutCleanup(t *testing.T) {
	original := executeSessionCreate
	cleanupCalled := false
	executeSessionCreate = func(_ BootLoader, _, _ []string, _, _ string) (sessionCreateResult, func(), error) {
		return sessionCreateResult{}, func() { cleanupCalled = true }, fmt.Errorf("create failed")
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}

func TestSessionListFormatsParentAndRunsCleanup(t *testing.T) {
	original := loadSessionList
	cleanupCalled := false
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return []sessionListEntry{{
			ID:        "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			Address:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			ParentID:  "parent-1234567890",
			ExpiresAt: "2026-05-15T00:00:00Z",
			Limit:     "unlimited",
			Status:    "expired",
		}}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "parent-1...")
	assert.Contains(t, out, "expired")
	assert.True(t, cleanupCalled)
}

func TestSessionListPropagatesLoadError(t *testing.T) {
	original := loadSessionList
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return nil, nil, fmt.Errorf("list failed")
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list failed")
	assert.Empty(t, out)
}

func TestSessionRevokeRequiresIDUnlessAllAndDoesNotExecute(t *testing.T) {
	original := executeSessionRevoke
	called := false
	executeSessionRevoke = func(_ BootLoader, _ bool, _ string) (string, func(), error) {
		called = true
		return "", nil, nil
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provide a session ID or use --all")
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestSessionRevokePropagatesExecutorError(t *testing.T) {
	original := executeSessionRevoke
	executeSessionRevoke = func(_ BootLoader, all bool, sessionID string) (string, func(), error) {
		assert.False(t, all)
		assert.Equal(t, "session-1", sessionID)
		return "", nil, fmt.Errorf("revoke failed")
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "session-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoke failed")
	assert.Empty(t, out)
}
