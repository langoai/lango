package chat

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlashCommandsStateReportingAndExitBranches(t *testing.T) {
	m := newTestModel()
	m.chatView.appendUser("hello")
	m.chatView.appendAssistant("world")
	originalSessionKey := m.sessionKey

	clearText := runSlashCommandText(t, cmdClear(m, ""))
	assert.Equal(t, "Chat cleared. New session started.", clearText)
	assert.Empty(t, m.chatView.entries)
	assert.Empty(t, strings.TrimSpace(m.chatView.viewport.View()))
	assert.NotEqual(t, originalSessionKey, m.sessionKey)
	assert.True(t, strings.HasPrefix(m.sessionKey, "tui-"))

	m.cfg.Agent.Provider = ""
	m.cfg.Agent.Model = ""
	modelText := runSlashCommandText(t, cmdModel(m, ""))
	assert.Equal(t, "Provider: (not configured)  Model: (auto)", modelText)

	modeListText := runSlashCommandText(t, cmdMode(m, ""))
	assert.Contains(t, modeListText, "Current mode: none")
	assert.Contains(t, modeListText, "Available modes:")

	unknownModeText := runSlashCommandText(t, cmdMode(m, "missing-mode"))
	assert.Equal(t, `Unknown mode "missing-mode". Type /mode for the list of available modes.`, unknownModeText)

	modeStore := newModeTestSessionStore()
	m.sessionStore = modeStore
	modeChangedText := runSlashCommandText(t, cmdMode(m, "debug"))
	assert.Equal(t, "Mode changed:  → debug", modeChangedText)
	require.NotNil(t, modeStore.sessions[m.sessionKey])
	assert.Equal(t, "debug", modeStore.sessions[m.sessionKey].Mode())

	assert.Equal(t, "Session totals: 0 input tokens, 0 output tokens, estimated cost —", runSlashCommandText(t, cmdCost(m, "")))

	m.sessionInputTokens = 123
	m.sessionOutputTokens = 45
	m.sessionCostUSD = 0.0042
	assert.Equal(t, "Session totals: 123 input tokens, 45 output tokens, estimated cost $0.0042", runSlashCommandText(t, cmdCost(m, "")))

	m.sessionCostUSD = 1.234
	assert.Equal(t, "Session totals: 123 input tokens, 45 output tokens, estimated cost $1.23", runSlashCommandText(t, cmdCost(m, "")))

	exitMsg := cmdExit(m, "")()
	assert.IsType(t, tea.QuitMsg{}, exitMsg)
}

func runSlashCommandText(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	require.NotNil(t, cmd)
	msg := cmd()
	sys, ok := msg.(SystemMsg)
	require.True(t, ok, "expected SystemMsg, got %T", msg)
	return sys.Text
}
