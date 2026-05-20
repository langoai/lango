package chat

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestHandleKeyCtrlDQuitsAndCancelsWithoutSubmitting(t *testing.T) {
	m := newTestModel()
	cancelled := false
	m.cancelFn = func() { cancelled = true }
	m.SetComposerValue("draft that must stay local")

	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlD})

	assert.True(t, cancelled)
	assert.True(t, m.quitting)
	assert.NotNil(t, cmd)
	assert.Equal(t, "draft that must stay local", m.ComposerValue())
}

func TestHandleKeyCancellingStateIgnoresNonQuitKeys(t *testing.T) {
	m := newTestModel()
	m.state = stateCancelling
	m.SetComposerValue("still cancelling")

	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	assert.Nil(t, cmd)
	assert.Equal(t, stateCancelling, m.state)
	assert.Equal(t, "still cancelling", m.ComposerValue())
}

func TestStreamingCtrlCTransitionsToCancellingAndPreservesDraft(t *testing.T) {
	m := newTestModel()
	m.state = stateStreaming
	m.SetComposerValue("next request")
	cancelled := false
	m.cancelFn = func() { cancelled = true }

	cmd := m.handleStreamingKey(tea.KeyMsg{Type: tea.KeyCtrlC})

	assert.True(t, cancelled)
	assert.Equal(t, stateCancelling, m.state)
	assert.Nil(t, cmd)
	assert.Equal(t, "next request", m.ComposerValue())
	assert.Empty(t, m.pendingRedirectInput)
}

func TestApprovingExpiredConfirmIsClearedBeforeKeyDispatch(t *testing.T) {
	m := newTestModel()
	m.state = stateApproving
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ID:       "apr-expired",
			ToolName: "exec",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: make(chan approval.ApprovalResponse, 1),
	}
	m.approval.confirmPending = true
	m.approval.confirmAction = "a"
	m.approval.confirmTime = time.Now().Add(-4 * time.Second)

	cmd := m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	assert.Nil(t, cmd)
	assert.False(t, m.approval.confirmPending)
	assert.Empty(t, m.approval.confirmAction)
	assert.NotNil(t, m.approval.pending)
	assert.Equal(t, stateApproving, m.state)
}

func TestApprovingCriticalSessionGrantSecondPressRespondsAlwaysAllow(t *testing.T) {
	respCh := make(chan approval.ApprovalResponse, 1)
	m := newTestModel()
	m.state = stateApproving
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ID:       "apr-session",
			ToolName: "exec",
		},
		ViewModel: approval.ApprovalViewModel{
			Risk: approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"},
		},
		Response: respCh,
	}

	first := m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Nil(t, first)
	assert.True(t, m.approval.confirmPending)
	assert.Equal(t, "s", m.approval.confirmAction)

	second := m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	for _, msg := range collectImmediateMsgs(second) {
		if msg != nil {
			updated, _ := m.Update(msg)
			m = updated.(*ChatModel)
		}
	}

	require.False(t, m.HasPendingApproval())
	assert.False(t, m.approval.confirmPending)
	assert.Equal(t, stateStreaming, m.state)
	resp := <-respCh
	assert.True(t, resp.Approved)
	assert.True(t, resp.AlwaysAllow)
	assert.Equal(t, "tui", resp.Provider)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemApproval, last.kind)
	assert.Contains(t, last.content, "Always allow enabled for exec")
}

func TestResetApprovalOwnerSharedPendingClearsLocalDialogState(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{ID: "shared-apr", ToolName: "exec"},
		},
	}
	m := newTestModelWithSharedPending(shared)
	m.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{ID: "local-apr", ToolName: "fs_write"},
	}
	m.approval.confirmPending = true
	m.approval.confirmAction = "a"
	m.approval.scrollOffset = 7
	m.approval.splitMode = true
	m.approval.diffCache = diffLineCache{content: "cached", width: 80, splitMode: true, lines: []string{"cached"}}

	m.resetApprovalOwner(&ApprovalRequestMsg{
		Request: approval.ApprovalRequest{ID: "incoming-apr", ToolName: "browser_search"},
	})

	assert.Nil(t, m.approval.pending)
	assert.False(t, m.approval.confirmPending)
	assert.Empty(t, m.approval.confirmAction)
	assert.Zero(t, m.approval.scrollOffset)
	assert.False(t, m.approval.splitMode)
	assert.Equal(t, "shared-apr", m.currentPendingApproval().Request.ID)
}

func TestSubmitCurrentInputSlashResetsDraftBeforeSetupGuard(t *testing.T) {
	m := newTestModel()
	m.cfg = incompleteSetupConfig()
	m.SetComposerValue(" /unknown-escrowToolsNonHubLifecycleOmitsOnChainFields3 ")

	cmd := m.submitCurrentInput(context.Background(), true)
	msgs := collectImmediateMsgs(cmd)

	assert.Empty(t, m.ComposerValue())
	assert.Equal(t, stateIdle, m.state)
	require.Len(t, msgs, 1)
	sys, ok := msgs[0].(SystemMsg)
	require.True(t, ok)
	assert.Contains(t, sys.Text, "Unknown command: /unknown-escrowtoolsnonhublifecycleomitsonchainfields3")
	assert.Empty(t, m.chatView.entries)
}

func TestSubmitCurrentInputSkipsSetupGuardWhenGuardDisabled(t *testing.T) {
	executor := &submitCaptureExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil)
	var submissions []string
	m := New(Deps{
		TurnRunner:       runner,
		Config:           incompleteSetupConfig(),
		SessionKey:       "escrowToolsNonHubLifecycleOmitsOnChainFields3-session",
		OnUserSubmission: func(_ string, input string) { submissions = append(submissions, input) },
	})
	m.SetComposerValue("run despite incomplete focused setup")

	cmd := m.submitCurrentInput(context.Background(), false)

	assert.NotNil(t, cmd)
	assert.Equal(t, stateStreaming, m.state)
	assert.True(t, m.pending.IsActive())
	assert.Empty(t, m.ComposerValue())
	assert.Equal(t, []string{"run despite incomplete focused setup"}, submissions)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemUser, last.kind)
	assert.Equal(t, "run despite incomplete focused setup", last.content)
	assert.Empty(t, executor.input, "turn command should not run until the returned tea.Cmd is executed")
}

func TestAgentSetupReadyAndSetupVisibilityBranches(t *testing.T) {
	var nilModel *ChatModel
	assert.False(t, nilModel.agentSetupReady())
	assert.False(t, nilModel.setupRequiredVisible())
	assert.False(t, nilModel.focusedSetupGuardEnabled())

	incomplete := newTestModel()
	incomplete.cfg = incompleteSetupConfig()
	incomplete.state = stateIdle
	assert.False(t, incomplete.agentSetupReady())
	assert.True(t, incomplete.focusedSetupGuardEnabled())
	assert.True(t, incomplete.setupRequiredVisible())

	incomplete.state = stateStreaming
	assert.False(t, incomplete.setupRequiredVisible())

	ready := newTestModel()
	ready.cfg = readyRemoteConfig()
	assert.True(t, ready.agentSetupReady())
	assert.False(t, ready.setupRequiredVisible())

	cockpitOwned := newTestModelWithSharedPending(&stubSharedPendingStore{})
	cockpitOwned.cfg = incompleteSetupConfig()
	assert.False(t, cockpitOwned.focusedSetupGuardEnabled())
	assert.False(t, cockpitOwned.setupRequiredVisible())
}

func TestReplayKeysAppliesInputKeysAndKeepsProcessingAfterCommandKeys(t *testing.T) {
	m := newTestModel()
	cmds := m.replayKeys([]tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
		{Type: tea.KeyRunes, Runes: []rune{'b'}},
	})

	assert.NotEmpty(t, cmds)
	assert.Equal(t, "ab", m.ComposerValue())

	cmds = m.replayKeys([]tea.KeyMsg{
		{Type: tea.KeyCtrlD},
		{Type: tea.KeyRunes, Runes: []rune{'c'}},
	})

	assert.NotEmpty(t, cmds)
	assert.True(t, m.quitting)
	assert.Equal(t, "abc", m.ComposerValue())
}

func TestDoneRedirectUsesSubmissionHookAndDoesNotRenderFailure(t *testing.T) {
	var submissions []string
	m := New(Deps{
		TurnRunner:       turnrunner.New(turnrunner.Config{}, &submitCaptureExecutor{}, submitTestSessionStore{}, nil),
		Config:           readyRemoteConfig(),
		SessionKey:       "escrowToolsNonHubLifecycleOmitsOnChainFields3-session",
		OnUserSubmission: func(sessionKey, input string) { submissions = append(submissions, sessionKey+":"+input) },
	})
	m.chatView.appendChunk("partial answer")
	m.pendingRedirectInput = "replacement prompt"

	updated, cmd := m.Update(DoneMsg{Result: turnrunner.Result{
		Outcome:     "failed",
		UserMessage: "should not be shown",
	}})
	m = updated.(*ChatModel)

	assert.NotNil(t, cmd)
	assert.Empty(t, m.pendingRedirectInput)
	assert.Equal(t, stateStreaming, m.state)
	assert.Equal(t, []string{"escrowToolsNonHubLifecycleOmitsOnChainFields3-session:replacement prompt"}, submissions)
	assert.True(t, m.pending.IsActive())

	var rendered []string
	for _, entry := range m.chatView.entries {
		rendered = append(rendered, entry.rawContent, entry.content)
	}
	joined := strings.Join(rendered, "\n")
	assert.Contains(t, joined, "partial answer")
	assert.Contains(t, joined, "replacement prompt")
	assert.NotContains(t, joined, "should not be shown")
}
