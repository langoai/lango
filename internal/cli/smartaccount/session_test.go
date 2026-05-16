package smartaccount

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountSessionCreate_WritesTextToCommandWriter(t *testing.T) {
	original := executeSessionCreate
	executeSessionCreate = func(_ BootLoader, _, _ []string, _, _ string) (sessionCreateResult, func(), error) {
		return sessionCreateResult{
			ID:        "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			Address:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			Targets:   []string{"0xaaaa"},
			Functions: []string{"0xa9059cbb"},
			Limit:     "1000000",
			ExpiresAt: "2026-05-15T00:00:00Z",
			CreatedAt: "2026-05-14T00:00:00Z",
		}, func() {}, nil
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--targets", "0x000000000000000000000000000000000000aaaa", "--duration", "24h")
	require.NoError(t, err)
	assert.Contains(t, out, "Session Key Created")
	assert.Contains(t, out, "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
}

func TestSmartAccountSessionCreate_WritesJSONToCommandWriter(t *testing.T) {
	original := executeSessionCreate
	executeSessionCreate = func(_ BootLoader, _, _ []string, _, _ string) (sessionCreateResult, func(), error) {
		return sessionCreateResult{
			ID:        "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			Address:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			Targets:   []string{"0xaaaa"},
			Functions: []string{"0xa9059cbb"},
			Limit:     "1000000",
			ExpiresAt: "2026-05-15T00:00:00Z",
			CreatedAt: "2026-05-14T00:00:00Z",
		}, func() {}, nil
	}
	t.Cleanup(func() { executeSessionCreate = original })

	cmd := sessionCreateCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--duration", "24h", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "a1b2c3d4-e5f6-7890-abcd-ef1234567890", decoded["id"])
	assert.Equal(t, "1000000", decoded["spendLimit"])
}

func TestSmartAccountSessionList_WritesTextToCommandWriter(t *testing.T) {
	original := loadSessionList
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return []sessionListEntry{
			{ID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890", Address: "0x1234abcd5678ef901234abcdef567890abcdef12", ExpiresAt: "2026-05-15T00:00:00Z", Limit: "unlimited", Status: "active"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "a1b2c3d4")
}

func TestSmartAccountSessionList_WritesJSONToCommandWriter(t *testing.T) {
	original := loadSessionList
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return []sessionListEntry{
			{ID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890", Address: "0x1234abcd5678ef901234abcdef567890abcdef12", ExpiresAt: "2026-05-15T00:00:00Z", Limit: "unlimited", Status: "active"},
		}, func() {}, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "active", decoded[0]["status"])
}

func TestSmartAccountSessionList_WritesEmptyStateToCommandWriter(t *testing.T) {
	original := loadSessionList
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		return nil, func() {}, nil
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No session keys found.")
}

func TestSmartAccountSessionList_InvalidOutputRejectsBeforeLoad(t *testing.T) {
	original := loadSessionList
	called := false
	loadSessionList = func(_ BootLoader) ([]sessionListEntry, func(), error) {
		called = true
		return nil, nil, assert.AnError
	}
	t.Cleanup(func() { loadSessionList = original })

	cmd := sessionListCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestSmartAccountSessionRevoke_WritesSingleSuccessToCommandWriter(t *testing.T) {
	original := executeSessionRevoke
	executeSessionRevoke = func(_ BootLoader, _ bool, sessionID string) (string, func(), error) {
		return "Session key " + sessionID + " revoked.", func() {}, nil
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	require.NoError(t, err)
	assert.Contains(t, out, "Session key a1b2c3d4-e5f6-7890-abcd-ef1234567890 revoked.")
}

func TestSmartAccountSessionRevoke_WritesAllSuccessToCommandWriter(t *testing.T) {
	original := executeSessionRevoke
	executeSessionRevoke = func(_ BootLoader, all bool, _ string) (string, func(), error) {
		assert.True(t, all)
		return "All active session keys revoked.", func() {}, nil
	}
	t.Cleanup(func() { executeSessionRevoke = original })

	cmd := sessionRevokeCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--all")
	require.NoError(t, err)
	assert.Contains(t, out, "All active session keys revoked.")
}
