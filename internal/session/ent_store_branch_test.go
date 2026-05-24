package session

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntStoreCloseAndNilMessageResolversAreNoops(t *testing.T) {
	store := &EntStore{}

	require.NoError(t, store.Close())

	content, err := store.resolveMessageContent(nil)
	require.NoError(t, err)
	require.Empty(t, content)

	toolCalls, err := store.resolveMessageToolCalls(nil)
	require.NoError(t, err)
	require.Nil(t, toolCalls)
}

func TestEntStoreCreateRejectsDuplicateSessionKey(t *testing.T) {
	store := newTestEntStore(t)

	require.NoError(t, store.Create(&Session{Key: "ent-store-branch-duplicate"}))

	err := store.Create(&Session{Key: "ent-store-branch-duplicate"})

	require.Error(t, err)
	require.ErrorIs(t, err, ErrDuplicateSession)
	require.Contains(t, err.Error(), `create session "ent-store-branch-duplicate"`)
}

func TestEntStoreUpdatePersistsOptionalFieldsAndMetadata(t *testing.T) {
	store := newTestEntStore(t)

	require.NoError(t, store.Create(&Session{Key: "ent-store-branch-update"}))

	err := store.Update(&Session{
		Key:         "ent-store-branch-update",
		AgentID:     "agent-branch",
		ChannelType: "slack",
		ChannelID:   "channel-123",
		Model:       "gpt-5.4",
		Metadata:    map[string]string{"workspace": "coverage"},
	})
	require.NoError(t, err)

	got, err := store.Get("ent-store-branch-update")
	require.NoError(t, err)
	require.Equal(t, "agent-branch", got.AgentID)
	require.Equal(t, "slack", got.ChannelType)
	require.Equal(t, "channel-123", got.ChannelID)
	require.Equal(t, "gpt-5.4", got.Model)
	require.Equal(t, map[string]string{"workspace": "coverage"}, got.Metadata)
}

func TestEntStoreCompactMessagesRejectsMissingAndOutOfRangeInputs(t *testing.T) {
	store := newTestEntStore(t)

	err := store.CompactMessages("missing-session", 0, "summary")
	require.Error(t, err)
	require.Contains(t, err.Error(), `get session "missing-session"`)

	require.NoError(t, store.Create(&Session{Key: "ent-store-branch-compact"}))
	require.NoError(t, store.AppendMessage("ent-store-branch-compact", Message{Role: "user", Content: "first"}))

	err = store.CompactMessages("ent-store-branch-compact", -1, "summary")
	require.Error(t, err)
	require.Contains(t, err.Error(), "compact index -1 out of range")

	err = store.CompactMessages("ent-store-branch-compact", 1, "summary")
	require.Error(t, err)
	require.Contains(t, err.Error(), "compact index 1 out of range")
}
