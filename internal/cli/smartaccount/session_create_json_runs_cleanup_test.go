package smartaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCreateJSONRunsCleanup(t *testing.T) {
	original := executeSessionCreate
	cleanupCalled := false
	executeSessionCreate = func(_ BootLoader, targets, functions []string, limit, duration string) (sessionCreateResult, func(), error) {
		assert.Equal(t, []string{"0x000000000000000000000000000000000000aaaa"}, targets)
		assert.Equal(t, []string{"0x095ea7b3", "0xa9059cbb"}, functions)
		assert.Equal(t, "123", limit)
		assert.Equal(t, "30m", duration)
		return sessionCreateResult{
			ID:        "session-abcdef12",
			Address:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			Targets:   []string{"0x000000000000000000000000000000000000aaaa"},
			Functions: []string{"0x095ea7b3", "0xa9059cbb"},
			Limit:     "123",
			ExpiresAt: "2026-05-15T00:30:00Z",
			CreatedAt: "2026-05-15T00:00:00Z",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd,
		"--targets", "0x000000000000000000000000000000000000aaaa",
		"--functions", "0x095ea7b3,0xa9059cbb",
		"--limit", "123",
		"--duration", "30m",
		"--output", "json",
	)

	require.NoError(t, err)
	assert.True(t, cleanupCalled)

	var decoded sessionCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "session-abcdef12", decoded.ID)
	assert.Equal(t, []string{"0x095ea7b3", "0xa9059cbb"}, decoded.Functions)
	assert.Equal(t, "123", decoded.Limit)
}

func TestSessionListEmptyJSONRunsCleanup(t *testing.T) {
	original := loadSessionList
	cleanupCalled := false
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return []sessionListEntry{}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")

	require.NoError(t, err)
	assert.True(t, cleanupCalled)

	var decoded []sessionListEntry
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Empty(t, decoded)
}

func TestSessionListHandlesShortTableValues(t *testing.T) {
	original := loadSessionList
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return []sessionListEntry{{
			ID:        "short",
			Address:   "0x1",
			ParentID:  "p",
			ExpiresAt: "2026-05-15T00:00:00Z",
			Limit:     "unlimited",
			Status:    "active",
		}}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "short")
	assert.Contains(t, out, "0x1")
	assert.Contains(t, out, "p")
	assert.NotContains(t, out, "short...")
}

func TestSessionRevokeRejectsAllWithSessionIDBeforeExecute(t *testing.T) {
	original := executeSessionRevoke
	called := false
	executeSessionRevoke = func(_ BootLoader, _ bool, _ string) (string, func(), error) {
		called = true
		return "", nil, nil
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--all", "session-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "use either --all or a session ID")
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestSessionRevokeRunsCleanupAfterSuccess(t *testing.T) {
	original := executeSessionRevoke
	cleanupCalled := false
	executeSessionRevoke = func(_ BootLoader, all bool, sessionID string) (string, func(), error) {
		assert.False(t, all)
		assert.Equal(t, "session-1", sessionID)
		return "Session key session-1 revoked.", func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "session-1")

	require.NoError(t, err)
	assert.Contains(t, out, "Session key session-1 revoked.")
	assert.True(t, cleanupCalled)
}

func TestSessionRevokePropagatesErrorWithoutCleanup(t *testing.T) {
	original := executeSessionRevoke
	cleanupCalled := false
	executeSessionRevoke = func(_ BootLoader, all bool, _ string) (string, func(), error) {
		assert.True(t, all)
		return "", func() { cleanupCalled = true }, fmt.Errorf("revoke all failed")
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--all")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoke all failed")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}
