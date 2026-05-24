package chat

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
)

func TestChatModelSessionKeyAndProgramAccessors(t *testing.T) {

	m := New(Deps{
		Config:     readyRemoteConfig(),
		SessionKey: "session-original",
	})
	assert.Equal(t, "session-original", m.SessionKey())

	p := tea.NewProgram(m)
	m.SetProgram(p)
	assert.Same(t, p, m.program)

	m.sessionKey = "session-after-clear"
	assert.Equal(t, "session-after-clear", m.SessionKey())
}

func TestChatModelInitIncludesTaskTickOnlyWithBackgroundManager(t *testing.T) {

	withoutBackground := New(Deps{Config: readyRemoteConfig(), SessionKey: "sess-1"})
	assert.Nil(t, withoutBackground.taskStripTick())
	require.NotNil(t, withoutBackground.Init())

	withBackground := New(Deps{
		Config:            readyRemoteConfig(),
		SessionKey:        "sess-2",
		BackgroundManager: background.NewManager(nil, nil, 1, time.Minute, nil),
	})
	assert.NotNil(t, withBackground.taskStripTick())
	require.NotNil(t, withBackground.Init())
}

func TestChatModelTaskStripTickRefreshesViewAndSchedulesNextTick(t *testing.T) {

	m := New(Deps{
		Config:            readyRemoteConfig(),
		SessionKey:        "sess-task",
		BackgroundManager: background.NewManager(nil, nil, 1, time.Minute, nil),
	})
	m.width = 100
	m.height = 24
	m.chatView.appendAssistant("stable transcript")

	updated, cmd := m.Update(TaskStripTickMsg(time.Now()))
	got := updated.(*ChatModel)

	assert.Same(t, m, got)
	assert.NotNil(t, cmd)
	require.NotEmpty(t, got.chatView.entries)
	assert.Equal(t, "stable transcript", got.chatView.entries[0].rawContent)
}

func TestChatModelIdleCtrlCWarnsThenQuitsOnDoublePress(t *testing.T) {

	m := newTestModel()

	cmd := m.handleIdleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	msgs := collectImmediateMsgs(cmd)
	require.Len(t, msgs, 1)
	sys, ok := msgs[0].(SystemMsg)
	require.True(t, ok)
	assert.Contains(t, sys.Text, "Press Ctrl+C again to quit")
	assert.False(t, m.quitting)

	cmd = m.handleIdleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
	assert.True(t, m.quitting)
}

func TestChatModelStreamingEnterQueuesRedirectAndCancelsCurrentTurn(t *testing.T) {

	m := newTestModel()
	m.state = stateStreaming
	m.SetComposerValue("follow up")
	cancelled := false
	m.cancelFn = func() { cancelled = true }
	m.chatView.appendChunk("partial")

	cmd := m.handleStreamingKey(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Nil(t, cmd)
	assert.True(t, cancelled)
	assert.Equal(t, "follow up", m.pendingRedirectInput)
	assert.Empty(t, m.ComposerValue())
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemStatus, last.kind)
	assert.Contains(t, last.content, "[interrupted]")
}

func TestChatModelApprovalKeysIgnoreMissingPendingAndSharedResolveFailure(t *testing.T) {

	m := newTestModel()
	m.state = stateApproving
	assert.Nil(t, m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}))

	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{ID: "apr-fail", ToolName: "exec"},
		},
		resolveOK: false,
	}
	m = newTestModelWithSharedPending(shared)
	m.state = stateApproving

	cmd := m.handleApprovingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, shared.resolveCount)
	require.NotEmpty(t, m.chatView.entries)
	last := m.chatView.entries[len(m.chatView.entries)-1]
	assert.Equal(t, itemStatus, last.kind)
	assert.Contains(t, last.content, "Approval resolution failed")
}

func TestGenerateAndTruncateSessionKey(t *testing.T) {

	key := generateSessionKey()
	assert.True(t, strings.HasPrefix(key, "tui-"))
	assert.Greater(t, len(key), len("tui-"))

	assert.Equal(t, "short-session", truncateSessionKey("short-session"))
	assert.Equal(t, "12345678901234567890...", truncateSessionKey("123456789012345678901"))
}
