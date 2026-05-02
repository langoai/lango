package proposal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time { return c.now }

func TestProposalRegistryCreateUpdateFromSameSource(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)}
	registry := NewRegistry(clock.Now)

	first, err := registry.Upsert(UpsertInput{
		SessionKey: "sess-1",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "suggestion-1"},
		Title:      "Apply retry rule",
		Summary:    "First summary",
		Reason:     "first reason",
		Confidence: 0.75,
		ExpiresAt:  clock.now.Add(30 * time.Minute),
	})
	require.NoError(t, err)

	clock.now = clock.now.Add(5 * time.Minute)

	second, err := registry.Upsert(UpsertInput{
		SessionKey: "sess-1",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "suggestion-1"},
		Title:      "Apply retry rule",
		Summary:    "Updated summary",
		Reason:     "updated reason",
		Confidence: 0.91,
		ExpiresAt:  clock.now.Add(45 * time.Minute),
	})
	require.NoError(t, err)

	assert.Equal(t, first.ProposalID, second.ProposalID)
	assert.Equal(t, "Updated summary", second.Summary)
	assert.Equal(t, "updated reason", second.Reason)
	assert.Equal(t, 0.91, second.Confidence)
	assert.Equal(t, ProposalStatusSuggested, second.Status)
	assert.True(t, second.UpdatedAt.After(second.CreatedAt))
}

func TestProposalRegistrySessionScoping(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)}
	registry := NewRegistry(clock.Now)

	_, err := registry.Upsert(UpsertInput{
		SessionKey: "sess-1",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "s-1"},
		Title:      "Proposal A",
		ExpiresAt:  clock.now.Add(time.Hour),
	})
	require.NoError(t, err)
	_, err = registry.Upsert(UpsertInput{
		SessionKey: "sess-2",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "s-1"},
		Title:      "Proposal B",
		ExpiresAt:  clock.now.Add(time.Hour),
	})
	require.NoError(t, err)

	sess1 := registry.ListBySession("sess-1")
	sess2 := registry.ListBySession("sess-2")

	require.Len(t, sess1, 1)
	require.Len(t, sess2, 1)
	assert.Equal(t, "Proposal A", sess1[0].Title)
	assert.Equal(t, "Proposal B", sess2[0].Title)
}

func TestProposalRegistryDismissBehavior(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)}
	registry := NewRegistry(clock.Now)

	created, err := registry.Upsert(UpsertInput{
		SessionKey: "sess-1",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "s-1"},
		Title:      "Dismiss me",
		ExpiresAt:  clock.now.Add(time.Hour),
	})
	require.NoError(t, err)

	clock.now = clock.now.Add(2 * time.Minute)
	dismissed, err := registry.Dismiss(created.ProposalID)
	require.NoError(t, err)

	assert.Equal(t, ProposalStatusDismissed, dismissed.Status)
	assert.Empty(t, registry.ListBySession("sess-1"))

	stored, ok := registry.GetByID(created.ProposalID)
	require.True(t, ok)
	assert.Equal(t, ProposalStatusDismissed, stored.Status)
}

func TestProposalRegistryExpirationPruneBehavior(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)}
	registry := NewRegistry(clock.Now)

	created, err := registry.Upsert(UpsertInput{
		SessionKey: "sess-1",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "s-expire"},
		Title:      "Expire me",
		ExpiresAt:  clock.now.Add(10 * time.Minute),
	})
	require.NoError(t, err)

	clock.now = clock.now.Add(11 * time.Minute)
	expired := registry.PruneExpired()
	assert.Equal(t, 1, expired)
	assert.Empty(t, registry.ListBySession("sess-1"))

	stored, ok := registry.GetByID(created.ProposalID)
	require.True(t, ok)
	assert.Equal(t, ProposalStatusExpired, stored.Status)
}

func TestProposalRegistryAcceptTransitionsOutOfVisibleActiveSet(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)}
	registry := NewRegistry(clock.Now)

	created, err := registry.Upsert(UpsertInput{
		SessionKey: "sess-1",
		Source:     ProposalSource{Kind: "proposed_learning", Ref: "s-accept"},
		Title:      "Accept me",
		ExpiresAt:  clock.now.Add(time.Hour),
	})
	require.NoError(t, err)

	clock.now = clock.now.Add(time.Minute)
	prepared, err := registry.MarkPrepared(created.ProposalID, PreparedBrief{
		SourceSummary:             "summary",
		Reason:                    "reason",
		SuggestedAcceptanceEffect: "effect",
		SupportingEvidence:        []string{"evidence"},
	})
	require.NoError(t, err)
	assert.Equal(t, ProposalStatusPrepared, prepared.Status)

	clock.now = clock.now.Add(time.Minute)
	accepted, err := registry.Accept(created.ProposalID)
	require.NoError(t, err)

	assert.Equal(t, ProposalStatusAccepted, accepted.Status)
	assert.Empty(t, registry.ListBySession("sess-1"))

	stored, ok := registry.GetByID(created.ProposalID)
	require.True(t, ok)
	assert.Equal(t, ProposalStatusAccepted, stored.Status)
	require.NotNil(t, stored.PreparedBrief)
	assert.Equal(t, "effect", stored.PreparedBrief.SuggestedAcceptanceEffect)
}
