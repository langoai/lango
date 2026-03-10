package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContributionTracker_RecordCommit(t *testing.T) {
	tracker := NewContributionTracker()

	tracker.RecordCommit("ws-1", "did:lango:agent-a", 1024)

	c := tracker.Get("ws-1", "did:lango:agent-a")
	require.NotNil(t, c)
	assert.Equal(t, "did:lango:agent-a", c.DID)
	assert.Equal(t, 1, c.Commits)
	assert.Equal(t, int64(1024), c.CodeBytes)
	assert.Equal(t, 0, c.Messages)
	assert.False(t, c.LastActive.IsZero())

	// Record a second commit and verify accumulation.
	tracker.RecordCommit("ws-1", "did:lango:agent-a", 512)
	c = tracker.Get("ws-1", "did:lango:agent-a")
	require.NotNil(t, c)
	assert.Equal(t, 2, c.Commits)
	assert.Equal(t, int64(1536), c.CodeBytes)
}

func TestContributionTracker_RecordMessage(t *testing.T) {
	tracker := NewContributionTracker()

	tracker.RecordMessage("ws-1", "did:lango:agent-b")

	c := tracker.Get("ws-1", "did:lango:agent-b")
	require.NotNil(t, c)
	assert.Equal(t, "did:lango:agent-b", c.DID)
	assert.Equal(t, 1, c.Messages)
	assert.Equal(t, 0, c.Commits)
	assert.False(t, c.LastActive.IsZero())

	// Record more messages.
	tracker.RecordMessage("ws-1", "did:lango:agent-b")
	tracker.RecordMessage("ws-1", "did:lango:agent-b")
	c = tracker.Get("ws-1", "did:lango:agent-b")
	require.NotNil(t, c)
	assert.Equal(t, 3, c.Messages)
}

func TestContributionTracker_Get(t *testing.T) {
	tests := []struct {
		give        string
		giveWS      string
		giveDID     string
		setup       func(*ContributionTracker)
		wantNil     bool
	}{
		{
			give:    "existing contribution",
			giveWS:  "ws-1",
			giveDID: "did:lango:agent-a",
			setup: func(tr *ContributionTracker) {
				tr.RecordCommit("ws-1", "did:lango:agent-a", 100)
			},
			wantNil: false,
		},
		{
			give:    "non-existent workspace",
			giveWS:  "ws-unknown",
			giveDID: "did:lango:agent-a",
			setup:   func(tr *ContributionTracker) {},
			wantNil: true,
		},
		{
			give:    "non-existent DID in existing workspace",
			giveWS:  "ws-1",
			giveDID: "did:lango:unknown",
			setup: func(tr *ContributionTracker) {
				tr.RecordCommit("ws-1", "did:lango:agent-a", 100)
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tracker := NewContributionTracker()
			tt.setup(tracker)

			c := tracker.Get(tt.giveWS, tt.giveDID)
			if tt.wantNil {
				assert.Nil(t, c)
			} else {
				assert.NotNil(t, c)
			}
		})
	}
}

func TestContributionTracker_List(t *testing.T) {
	tracker := NewContributionTracker()

	// Empty workspace returns nil.
	list := tracker.List("ws-1")
	assert.Nil(t, list)

	// Add contributions from multiple agents.
	tracker.RecordCommit("ws-1", "did:lango:agent-a", 100)
	tracker.RecordMessage("ws-1", "did:lango:agent-b")
	tracker.RecordCommit("ws-1", "did:lango:agent-c", 200)

	list = tracker.List("ws-1")
	assert.Len(t, list, 3)

	// Verify all DIDs are present.
	dids := make(map[string]bool, len(list))
	for _, c := range list {
		dids[c.DID] = true
	}
	assert.True(t, dids["did:lango:agent-a"])
	assert.True(t, dids["did:lango:agent-b"])
	assert.True(t, dids["did:lango:agent-c"])
}

func TestContributionTracker_Remove(t *testing.T) {
	tracker := NewContributionTracker()

	tracker.RecordCommit("ws-1", "did:lango:agent-a", 100)
	tracker.RecordMessage("ws-1", "did:lango:agent-b")

	require.Len(t, tracker.List("ws-1"), 2)

	tracker.Remove("ws-1")

	assert.Nil(t, tracker.List("ws-1"))
	assert.Nil(t, tracker.Get("ws-1", "did:lango:agent-a"))
}
