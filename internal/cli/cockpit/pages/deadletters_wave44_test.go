package pages

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/postadjudicationstatus"
)

func TestDeadLettersPage_Wave44LifecycleAndMissingDependencies(t *testing.T) {
	page := NewDeadLettersPage(nil, nil)

	assert.Nil(t, page.Init())
	page.Deactivate()

	page.selectedID = "tx-missing-detail"
	detailMsg, ok := page.loadSelectedDetail()().(deadLetterDetailLoadedMsg)
	require.True(t, ok)
	assert.Equal(t, "tx-missing-detail", detailMsg.transactionID)
	require.Error(t, detailMsg.err)
	assert.Contains(t, detailMsg.err.Error(), "dead-letter detail function not configured")

	page.detail = &postadjudicationstatus.TransactionStatus{CanRetry: true}
	assert.Nil(t, page.retrySelected())
}

func TestDeadLettersPage_Wave44ReverseFilterCyclingAndBackspaceFields(t *testing.T) {
	page := NewDeadLettersPage(nil, nil)

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyLeft},
		{Type: tea.KeyRunes, Runes: []rune("[")},
		{Type: tea.KeyRunes, Runes: []rune(",")},
		{Type: tea.KeyRunes, Runes: []rune(";")},
	} {
		updated, cmd := page.Update(keyMsg)
		page = updated.(*DeadLettersPage)
		assert.Nil(t, cmd)
	}

	assert.Equal(t, deadLetterAdjudicationRefund, page.adjudicationDraft)
	assert.Equal(t, deadLetterSubtypeDeadLettered, page.subtypeDraft)
	assert.Equal(t, deadLetterFamilyDeadLetter, page.familyDraft)
	assert.Equal(t, deadLetterFamilyDeadLetter, page.anyMatchFamilyDraft)

	page.queryDraft = "query"
	page.manualReplayActorDraft = "actor"
	page.deadLetteredAfterDraft = "after"
	page.deadLetteredBeforeDraft = "before"
	page.deadLetterReasonQueryDraft = "reason"
	page.latestDispatchReferenceDraft = "dispatch"

	for _, field := range []deadLetterTextField{
		deadLetterTextFieldQuery,
		deadLetterTextFieldManualReplayActor,
		deadLetterTextFieldDeadLetteredAfter,
		deadLetterTextFieldDeadLetteredBefore,
		deadLetterTextFieldDeadLetterReasonQuery,
		deadLetterTextFieldLatestDispatchReference,
	} {
		page.activeTextField = field
		updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		page = updated.(*DeadLettersPage)
		assert.Nil(t, cmd)
	}

	assert.Equal(t, "quer", page.queryDraft)
	assert.Equal(t, "acto", page.manualReplayActorDraft)
	assert.Equal(t, "afte", page.deadLetteredAfterDraft)
	assert.Equal(t, "befor", page.deadLetteredBeforeDraft)
	assert.Equal(t, "reaso", page.deadLetterReasonQueryDraft)
	assert.Equal(t, "dispatc", page.latestDispatchReferenceDraft)
	assert.Equal(t, "", trimLastByte(""))
}

func TestDeadLettersPage_Wave44RetryFollowUpRefreshFailuresAndFallbacks(t *testing.T) {
	t.Run("refresh failure records sanitized follow-up", func(t *testing.T) {
		page := NewDeadLettersPage(nil, nil)
		page.retryFollowUpPending = true
		page.retryFollowUpID = "tx-\x1b[31m1\nx"

		updated, cmd := page.Update(deadLettersLoadedMsg{err: errors.New("refresh\nfailed")})
		page = updated.(*DeadLettersPage)

		assert.Nil(t, cmd)
		assert.False(t, page.retryFollowUpPending)
		assert.Contains(t, page.statusMsg, "Retry request accepted for tx-1 x")
		assert.Contains(t, page.statusMsg, "refresh failed: refresh failed")
		assert.NotContains(t, page.statusMsg, "\x1b")
		assert.Empty(t, page.items)
		assert.Nil(t, page.detail)
	})

	t.Run("detail failure records follow-up note", func(t *testing.T) {
		page := NewDeadLettersPage(nil, nil)
		page.selectedID = "tx-1"
		page.retryFollowUpPending = true
		page.retryFollowUpID = "tx-1"

		updated, cmd := page.Update(deadLetterDetailLoadedMsg{
			transactionID: "tx-1",
			err:           errors.New("detail failed"),
		})
		page = updated.(*DeadLettersPage)

		assert.Nil(t, cmd)
		assert.False(t, page.retryFollowUpPending)
		assert.Contains(t, page.statusMsg, "latest status load failed: detail failed")
		assert.Nil(t, page.detail)
	})

	t.Run("loaded stale detail is ignored", func(t *testing.T) {
		page := NewDeadLettersPage(nil, nil)
		page.selectedID = "tx-current"

		updated, cmd := page.Update(deadLetterDetailLoadedMsg{
			transactionID: "tx-stale",
			status:        postadjudicationstatus.TransactionStatus{CanRetry: true},
		})
		page = updated.(*DeadLettersPage)

		assert.Nil(t, cmd)
		assert.Nil(t, page.detail)
	})
}

func TestDeadLettersPage_Wave44HelperBranches(t *testing.T) {
	assert.False(t, isDeadLetterQueryInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}}))
	assert.Equal(t, "inactive", deadLetterStatusLabel(postadjudicationstatus.DeadLetterBacklogEntry{}))
	assert.Equal(t, "-3", formatDeadLetterDelta(-3))
	assert.True(t, parseDeadLetterTimestamp("not-rfc3339").IsZero())

	page := NewDeadLettersPage(nil, nil)
	page.retryConfirmID = "tx-1"
	page.selectedID = "tx-1"
	page.detail = &postadjudicationstatus.TransactionStatus{CanRetry: true}
	assert.Equal(t, "disabled", page.retryActionLabel())

	page.retryFn = func(context.Context, string) error { return nil }
	assert.Equal(t, "still present in dead-letter backlog (subtype retry-scheduled, family retry, attempt 2)", page.describeRetryFollowUp(postadjudicationstatus.TransactionStatus{
		IsDeadLettered: true,
		RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
			LatestStatusSubtype:       "retry-scheduled",
			LatestStatusSubtypeFamily: "retry",
			LatestRetryAttempt:        2,
		},
	}))
	assert.Equal(t, "task unknown", page.describeRetryFollowUp(postadjudicationstatus.TransactionStatus{
		IsDeadLettered:       true,
		LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{},
	}))
}
