package chat

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestSharedApprovalResolveFailureKeepsPendingAndShowsStatus(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-networkModuleMetadataAndEnablementBranches5",
				ToolName: "exec",
				Summary:  "Run command",
			},
		},
		resolveOK: false,
	}
	m := newTestModelWithSharedPending(shared)
	m.state = stateApproving

	cmd := m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, shared.resolveCount)
	assert.Equal(t, "apr-networkModuleMetadataAndEnablementBranches5", shared.lastResolvedID)
	assert.False(t, shared.lastResponse.Approved)
	assert.Equal(t, "tui", shared.lastResponse.Provider)
	assert.Same(t, shared.latest, m.currentPendingApproval())
	assert.Equal(t, stateApproving, m.state)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemStatus, last.kind)
	assert.Contains(t, last.content, "Approval resolution failed for exec")
}

func TestSharedApprovalResolveAdvancesToNextPending(t *testing.T) {
	next := &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ID:       "apr-next",
			ToolName: "fs_write",
		},
	}
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-current",
				ToolName: "exec",
			},
		},
		next:      next,
		resolveOK: true,
	}
	m := newTestModelWithSharedPending(shared)
	m.state = stateApproving

	_ = m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	assert.Equal(t, 1, shared.resolveCount)
	assert.Equal(t, "apr-current", shared.lastResolvedID)
	assert.True(t, shared.lastResponse.Approved)
	assert.False(t, shared.lastResponse.AlwaysAllow)
	assert.Equal(t, next, m.currentPendingApproval())
	assert.Equal(t, stateApproving, m.state)
}

func TestDoneMsgFlushesTurnTokenUsageCommandAndResetsCounters(t *testing.T) {
	m := newTestModel()
	m.turnActive = true
	m.turnInputTokens = 11
	m.turnOutputTokens = 7
	m.turnCacheTokens = 3
	m.turnCostUSD = 0.015

	updated, cmd := m.Update(DoneMsg{Result: turnrunner.Result{Outcome: "success"}})
	m = updated.(*ChatModel)

	require.False(t, m.turnActive)
	assert.Zero(t, m.turnInputTokens)
	assert.Zero(t, m.turnOutputTokens)
	assert.Zero(t, m.turnCacheTokens)
	assert.Zero(t, m.turnCostUSD)

	msgs := collectImmediateMsgs(cmd)
	require.Len(t, msgs, 1)
	usage, ok := msgs[0].(TurnTokenUsageMsg)
	require.True(t, ok, "expected TurnTokenUsageMsg, got %T", msgs[0])
	assert.Equal(t, int64(11), usage.InputTokens)
	assert.Equal(t, int64(7), usage.OutputTokens)
	assert.Equal(t, int64(18), usage.TotalTokens)
	assert.Equal(t, int64(3), usage.CacheTokens)
	assert.Equal(t, 0.015, usage.EstimatedCostUSD)
}

func TestBudgetWarningRendersDeterministicStatus(t *testing.T) {
	m := newTestModel()

	updated, cmd := m.Update(BudgetWarningMsg{Used: 3, Max: 4})
	m = updated.(*ChatModel)

	assert.Nil(t, cmd)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemStatus, last.kind)
	rendered := strings.Join([]string{last.content, last.rawContent}, "\n")
	assert.Contains(t, rendered, "Delegation budget")
	assert.Contains(t, rendered, "3/4")
	assert.Contains(t, rendered, "75%")
}
