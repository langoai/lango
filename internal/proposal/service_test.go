package proposal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spyRegistry struct {
	inner *Registry
	calls []string
}

func newSpyRegistry(nowFn func() time.Time) *spyRegistry {
	return &spyRegistry{inner: NewRegistry(nowFn)}
}

func (s *spyRegistry) Upsert(in UpsertInput) (*Proposal, error) {
	s.calls = append(s.calls, "upsert")
	return s.inner.Upsert(in)
}

func (s *spyRegistry) GetByID(proposalID string) (Proposal, bool) {
	return s.inner.GetByID(proposalID)
}

func (s *spyRegistry) MarkPreparing(proposalID string) (*Proposal, error) {
	s.calls = append(s.calls, "mark-preparing")
	return s.inner.MarkPreparing(proposalID)
}

func (s *spyRegistry) MarkPrepared(proposalID string, brief PreparedBrief) (*Proposal, error) {
	s.calls = append(s.calls, "mark-prepared")
	return s.inner.MarkPrepared(proposalID, brief)
}

func (s *spyRegistry) Dismiss(proposalID string) (*Proposal, error) {
	s.calls = append(s.calls, "dismiss")
	return s.inner.Dismiss(proposalID)
}

func (s *spyRegistry) Accept(proposalID string) (*Proposal, error) {
	s.calls = append(s.calls, "accept")
	return s.inner.Accept(proposalID)
}

func (s *spyRegistry) RestorePrepared(proposalID string) (*Proposal, error) {
	s.calls = append(s.calls, "restore-prepared")
	return s.inner.RestorePrepared(proposalID)
}

func (s *spyRegistry) PruneExpired() int {
	s.calls = append(s.calls, "prune")
	return s.inner.PruneExpired()
}

func TestProposalServicePerformsSuggestedPreparingPrepared(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC)}
	registry := newSpyRegistry(clock.Now)
	service := NewService(registry, NewDeterministicPreparer())

	proposal, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "suggestion-1",
		Pattern:      "retry timeout",
		ProposedRule: "Use bounded retry",
		Confidence:   0.75,
		Rationale:    "Repeated timeout failures benefited from bounded retry.",
		ExpiresAt:    clock.now.Add(30 * time.Minute),
	})
	require.NoError(t, err)
	require.NotNil(t, proposal)

	assert.Equal(t, []string{"prune", "upsert", "mark-preparing", "mark-prepared"}, registry.calls)
	assert.Equal(t, ProposalStatusPrepared, proposal.Status)
	require.NotNil(t, proposal.PreparedBrief)
	assert.Equal(t, "Repeated timeout failures benefited from bounded retry.", proposal.PreparedBrief.Reason)
}

func TestProposalServiceDismissAndAcceptUseRegistry(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC)}
	registry := newSpyRegistry(clock.Now)
	service := NewService(registry, NewDeterministicPreparer())

	first, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "dismiss-me",
		Pattern:      "pattern-a",
		ProposedRule: "Rule A",
		ExpiresAt:    clock.now.Add(time.Hour),
	})
	require.NoError(t, err)
	registry.calls = nil

	dismissed, err := service.Dismiss(context.Background(), first.ProposalID)
	require.NoError(t, err)
	assert.Equal(t, []string{"dismiss"}, registry.calls)
	assert.Equal(t, ProposalStatusDismissed, dismissed.Status)

	second, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "accept-me",
		Pattern:      "pattern-b",
		ProposedRule: "Rule B",
		ExpiresAt:    clock.now.Add(time.Hour),
	})
	require.NoError(t, err)
	registry.calls = nil

	accepted, err := service.Accept(context.Background(), second.ProposalID)
	require.NoError(t, err)
	assert.Equal(t, []string{"prune", "accept"}, registry.calls)
	assert.Equal(t, ProposalStatusAccepted, accepted.Status)
}

func TestProposalServiceRestorePreparedReturnsAcceptedProposalToVisibleState(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC)}
	registry := newSpyRegistry(clock.Now)
	service := NewService(registry, NewDeterministicPreparer())

	prepared, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "restore-me",
		Pattern:      "pattern-r",
		ProposedRule: "Rule restore",
		Rationale:    "Keep prepared context",
		ExpiresAt:    clock.now.Add(time.Hour),
	})
	require.NoError(t, err)
	registry.calls = nil

	accepted, err := service.Accept(context.Background(), prepared.ProposalID)
	require.NoError(t, err)
	assert.Equal(t, ProposalStatusAccepted, accepted.Status)

	restored, err := service.RestorePrepared(context.Background(), prepared.ProposalID)
	require.NoError(t, err)
	assert.Equal(t, []string{"prune", "accept", "restore-prepared"}, registry.calls)
	assert.Equal(t, ProposalStatusPrepared, restored.Status)
	require.NotNil(t, restored.PreparedBrief)
	assert.NotEmpty(t, restored.PreparedBrief.SourceSummary)
}

func TestProposalServiceAcceptRejectsExpiredProposal(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC)}
	registry := newSpyRegistry(clock.Now)
	service := NewService(registry, NewDeterministicPreparer())

	prepared, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-1",
		SuggestionID: "expired-accept",
		Pattern:      "pattern-expired",
		ProposedRule: "Rule expired",
		ExpiresAt:    clock.now.Add(2 * time.Minute),
	})
	require.NoError(t, err)

	clock.now = clock.now.Add(3 * time.Minute)
	registry.calls = nil

	accepted, err := service.Accept(context.Background(), prepared.ProposalID)
	require.Error(t, err)
	assert.Nil(t, accepted)
	assert.Equal(t, []string{"prune"}, registry.calls)

	stored, ok := registry.GetByID(prepared.ProposalID)
	require.True(t, ok)
	assert.Equal(t, ProposalStatusExpired, stored.Status)
}

func TestProposalServiceRequiresRegistryAcrossEntryPoints(t *testing.T) {
	t.Parallel()

	service := &Service{preparer: NewDeterministicPreparer()}

	upserted, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-missing-registry",
		SuggestionID: "suggestion-1",
		Pattern:      "retry timeout",
		ProposedRule: "Use bounded retry",
	})
	require.Error(t, err)
	assert.Nil(t, upserted)
	assert.Contains(t, err.Error(), "upsert learning suggestion")
	assert.Contains(t, err.Error(), "proposal registry is required")

	accepted, err := service.Accept(context.Background(), "proposal-1")
	require.Error(t, err)
	assert.Nil(t, accepted)
	assert.Contains(t, err.Error(), "accept proposal")
	assert.Contains(t, err.Error(), "proposal registry is required")

	pruned, err := service.PruneExpired(context.Background())
	require.Error(t, err)
	assert.Zero(t, pruned)
	assert.Contains(t, err.Error(), "prune proposals")
	assert.Contains(t, err.Error(), "proposal registry is required")
}

func TestProposalServiceRequiresPreparerForLearningSuggestion(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC)}
	service := &Service{
		registry: newSpyRegistry(clock.Now),
	}

	upserted, err := service.UpsertLearningSuggestion(context.Background(), LearningSuggestionSource{
		SessionKey:   "sess-missing-preparer",
		SuggestionID: "suggestion-1",
		Pattern:      "retry timeout",
		ProposedRule: "Use bounded retry",
	})
	require.Error(t, err)
	assert.Nil(t, upserted)
	assert.Contains(t, err.Error(), "proposal preparer is required")
}
