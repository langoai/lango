package chat

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestWave40SetComposerValuePreservesEdgeCaseInput(t *testing.T) {
	var nilModel *ChatModel
	require.NotPanics(t, func() {
		nilModel.SetComposerValue("ignored")
	})

	m := newTestModel()
	m.SetComposerValue("  keep surrounding whitespace  \nsecond line")
	assert.Equal(t, "  keep surrounding whitespace  \nsecond line", m.ComposerValue())

	m.state = stateApproving
	m.SetComposerValue("")
	assert.Empty(t, m.ComposerValue())
	assert.False(t, m.CanStartTurnFromComposer())
}

func TestWave40PendingApprovalKeyNilSharedAndFullscreenBranches(t *testing.T) {
	m := newTestModel()
	approveKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	assert.False(t, m.CanHandlePendingApprovalKey(approveKey))
	assert.Nil(t, m.HandlePendingApprovalKey(approveKey))

	sharedEmpty := newTestModelWithSharedPending(&stubSharedPendingStore{})
	assert.False(t, sharedEmpty.CanHandlePendingApprovalKey(approveKey))
	assert.Nil(t, sharedEmpty.HandlePendingApprovalKey(approveKey))

	fullscreen := newTestModel()
	fullscreen.state = stateStreaming
	fullscreen.approval.pending = &ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ID:       "apr-wave40",
			ToolName: "fs_write",
			Summary:  "Review generated diff",
		},
		ViewModel: approval.ApprovalViewModel{
			Tier:        approval.TierFullscreen,
			DiffContent: strings.Repeat("+ changed\n", 12),
		},
		Response: make(chan approval.ApprovalResponse, 1),
	}

	assert.True(t, fullscreen.CanHandlePendingApprovalKey(tea.KeyMsg{Type: tea.KeyDown}))
	assert.Nil(t, fullscreen.HandlePendingApprovalKey(tea.KeyMsg{Type: tea.KeyDown}))
	assert.Equal(t, stateApproving, fullscreen.state)
	assert.Equal(t, 3, fullscreen.approval.scrollOffset)

	assert.True(t, fullscreen.CanHandlePendingApprovalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}))
	assert.Nil(t, fullscreen.HandlePendingApprovalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}))
	assert.Zero(t, fullscreen.approval.scrollOffset)

	assert.True(t, fullscreen.CanHandlePendingApprovalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}))
	assert.Nil(t, fullscreen.HandlePendingApprovalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}))
	assert.True(t, fullscreen.approval.splitMode)
}

func TestWave40UpdateResizeTickAndNonKeyBranchesWithoutProgram(t *testing.T) {
	m := newTestModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 96, Height: 31})
	m = updated.(*ChatModel)
	assert.Nil(t, cmd)
	assert.Equal(t, 96, m.width)
	assert.Equal(t, 31, m.height)
	assert.Equal(t, 96, m.chatView.viewport.Width)
	assert.GreaterOrEqual(t, m.chatView.viewport.Height, 3)

	updated, cmd = m.Update(TaskStripTickMsg(time.Now()))
	m = updated.(*ChatModel)
	assert.Nil(t, cmd)

	updated, cmd = m.Update(PendingIndicatorTickMsg(time.Now()))
	m = updated.(*ChatModel)
	assert.Nil(t, cmd)

	m.pending.Activate()
	updated, cmd = m.Update(PendingIndicatorTickMsg(time.Now()))
	m = updated.(*ChatModel)
	assert.NotNil(t, cmd)

	m.state = stateStreaming
	m.chatView.cursorTickActive = true
	m.chatView.showCursor = true
	updated, cmd = m.Update(CursorTickMsg(time.Now()))
	m = updated.(*ChatModel)
	assert.NotNil(t, cmd)
	assert.False(t, m.chatView.showCursor)

	m.state = stateIdle
	updated, cmd = m.Update(CursorTickMsg(time.Now()))
	m = updated.(*ChatModel)
	assert.Nil(t, cmd)
	assert.False(t, m.chatView.cursorTickActive)

	beforeEntries := len(m.chatView.entries)
	updated, cmd = m.Update(struct{ marker string }{marker: "ignored"})
	m = updated.(*ChatModel)
	assert.Nil(t, cmd)
	assert.Len(t, m.chatView.entries, beforeEntries)
}

func TestWave40SubmitCurrentInputNoopAndSetupGuardBranches(t *testing.T) {
	m := newTestModel()
	m.SetComposerValue(" \n\t ")
	assert.Nil(t, m.submitCurrentInput(context.Background(), true))
	assert.Empty(t, strings.TrimSpace(m.ComposerValue()))
	assert.Equal(t, stateIdle, m.state)

	m = newTestModel()
	m.cfg = incompleteSetupConfig()
	m.SetComposerValue("run blocked by setup guard")
	cmd := m.submitCurrentInput(context.Background(), true)
	require.NotNil(t, cmd)
	assert.Equal(t, "run blocked by setup guard", m.ComposerValue())
	assert.Equal(t, stateIdle, m.state)
	require.Len(t, m.chatView.entries, 1)
	assert.Equal(t, itemStatus, m.chatView.entries[0].kind)
	assert.Contains(t, m.chatView.entries[0].content, "lango onboard")
	assert.Empty(t, collectImmediateMsgs(cmd))
}

func TestWave40SubmitCmdNilParentReturnsDoneAndTracksTurnState(t *testing.T) {
	executor := &submitCaptureExecutor{}
	m := New(Deps{
		TurnRunner: turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil),
		Config:     readyRemoteConfig(),
		SessionKey: "wave40-session",
	})

	cmd := m.submitCmd(nil, "direct submit")
	require.NotNil(t, cmd)
	require.NotNil(t, m.runCtx)
	require.NotNil(t, m.cancelFn)
	assert.True(t, m.turnActive)

	msgs := collectImmediateMsgs(cmd)
	require.Len(t, msgs, 1)
	done, ok := msgs[0].(DoneMsg)
	require.True(t, ok)
	assert.Equal(t, "ok", done.Result.ResponseText)
	assert.Equal(t, "wave40-session", executor.sessionID)
	assert.Equal(t, "direct submit", executor.input)
	assert.NotNil(t, executor.ctx)
}

func TestWave40ViewFooterAndSetupBranches(t *testing.T) {
	m := newTestModel()
	m.quitting = true
	assert.Empty(t, m.View())

	m = newTestModel()
	m.width = 0
	m.height = 0
	assert.Contains(t, m.View(), "Waiting for terminal size")

	m = newTestModel()
	m.state = stateApproving
	footer := ansi.Strip(renderFooterWithSetup(m.input, m.state, 80, false))
	assert.NotContains(t, footer, defaultComposerPlaceholder)

	m = newTestModel()
	m.cfg = incompleteSetupConfig()
	m.state = stateIdle
	m.width = 100
	m.height = 24
	m.recalcLayout()
	parts := m.RenderParts()
	out := ansi.Strip(strings.Join([]string{parts.Header, parts.TurnStrip, parts.Footer}, "\n"))
	assert.Contains(t, out, "Setup Required")
	assert.Contains(t, out, setupRequiredGuidance)
	assert.NotContains(t, out, "Enter send")
}
