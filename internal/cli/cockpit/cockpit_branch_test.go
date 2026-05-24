package cockpit

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/cockpit/sidebar"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

type cockpitBranchMsg string

type commandPage struct {
	mockPage
	msg cockpitBranchMsg
}

func (p *commandPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p.updates = append(p.updates, msg)
	if p.msg == "" {
		return p, nil
	}
	return p, func() tea.Msg { return p.msg }
}

func TestNilModelAccessorsReturnZeroValues(t *testing.T) {
	var m *Model

	assert.Nil(t, m.Pages())
	assert.Equal(t, sidebar.Model{}, m.Sidebar())
	assert.Nil(t, m.ChatModel())
}

func TestAccessorsExposeRegisteredPagesAndSidebarState(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}

	m.RegisterPage(PageTools, toolsPage)

	assert.Same(t, toolsPage, m.Pages()[PageTools])
	assert.False(t, m.Sidebar().IsDisabled(PageTools.String()))
}

func TestInitActivatesRegisteredNonChatPage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.activePage = PageMissionControl
	missionPage := &mockPage{title: "Mission Control"}
	m.RegisterPage(PageMissionControl, missionPage)

	m.Init()

	assert.True(t, missionPage.activated)
}

func TestInitSkipsPageActivationForChatAndMissingPage(t *testing.T) {
	t.Run("chat", func(t *testing.T) {
		mock := &mockChild{}
		m := newTestModel(mock)
		chatPage := &mockPage{title: "Chat page should not be activated"}
		m.RegisterPage(PageChat, chatPage)

		m.Init()

		assert.False(t, chatPage.activated)
	})

	t.Run("missing page", func(t *testing.T) {
		mock := &mockChild{}
		m := newTestModel(mock)
		m.activePage = PageMissionControl

		require.NotPanics(t, func() {
			m.Init()
		})
	})
}

func TestContextTickPushesChannelAndRuntimeTrackersToPanel(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)
	channelTracker := NewChannelTracker(nil)
	channelTracker.SeedChannel("  discord\nops  ", true)
	runtimeTracker := NewRuntimeTracker(nil, nil, "sess-1")
	runtimeTracker.StartTurn()
	runtimeTracker.RecordDelegation("  helper\nagent  ")

	m.SetChannelTracker(channelTracker)
	m.SetRuntimeTracker(runtimeTracker)

	m.Update(contextTickMsg(time.Now()))

	require.Len(t, m.contextPanel.channelStatuses, 1)
	assert.Equal(t, "discord ops", m.contextPanel.channelStatuses[0].Name)
	assert.True(t, m.contextPanel.channelStatuses[0].Connected)
	assert.Equal(t, "helper agent", m.contextPanel.runtimeStat.ActiveAgent)
	assert.Equal(t, 1, m.contextPanel.runtimeStat.DelegationCount)
	assert.True(t, m.contextPanel.runtimeStat.IsRunning)
}

func TestWindowSizeForwardsPageCommandAndNoSidebarWidth(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)
	m.sidebarVisible = false
	m.contextVisible = true
	m.contextPanel.SetVisible(true)
	toolsPage := &commandPage{
		mockPage: mockPage{title: "Tools"},
		msg:      cockpitBranchMsg("tools-resized"),
	}
	m.RegisterPage(PageTools, toolsPage)

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	require.NotNil(t, cmd)
	msgs := collectCockpitImmediateMsgs(cmd)
	assert.Contains(t, msgs, cockpitBranchMsg("tools-resized"))
	require.NotEmpty(t, mock.updates)
	size, ok := mock.updates[0].(tea.WindowSizeMsg)
	require.True(t, ok)
	assert.Equal(t, 120-theme.ContextPanelWidth, size.Width)
}

func TestGlobalPageKeysSwitchPagesEvenWhenSidebarFocused(t *testing.T) {
	tests := []struct {
		name     string
		bind     func(*keyMap, string)
		target   PageID
		register bool
	}{
		{
			name:   "chat",
			bind:   func(km *keyMap, keyName string) { km.Page1 = key.NewBinding(key.WithKeys(keyName)) },
			target: PageChat,
		},
		{
			name:     "settings",
			bind:     func(km *keyMap, keyName string) { km.Page2 = key.NewBinding(key.WithKeys(keyName)) },
			target:   PageSettings,
			register: true,
		},
		{
			name:     "tools",
			bind:     func(km *keyMap, keyName string) { km.Page3 = key.NewBinding(key.WithKeys(keyName)) },
			target:   PageTools,
			register: true,
		},
		{
			name:     "status",
			bind:     func(km *keyMap, keyName string) { km.Page4 = key.NewBinding(key.WithKeys(keyName)) },
			target:   PageStatus,
			register: true,
		},
		{
			name:     "tasks",
			bind:     func(km *keyMap, keyName string) { km.Page5 = key.NewBinding(key.WithKeys(keyName)) },
			target:   PageTasks,
			register: true,
		},
		{
			name:     "approvals",
			bind:     func(km *keyMap, keyName string) { km.Page6 = key.NewBinding(key.WithKeys(keyName)) },
			target:   PageApprovals,
			register: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockChild{}
			m := newTestModel(mock)
			m.sidebarFocused = true
			m.sidebar.SetFocused(true)
			toolsPage := &mockPage{title: "Tools"}
			m.RegisterPage(PageTools, toolsPage)
			if tt.target == PageTools {
				m.activePage = PageChat
			} else {
				m.activePage = PageTools
			}
			if tt.register && tt.target != PageTools {
				m.RegisterPage(tt.target, &mockPage{title: tt.target.String()})
			}
			keyName := "x"
			tt.bind(&m.keymap, keyName)

			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyName)})

			assert.Equal(t, tt.target, m.activePage)
			assert.False(t, m.sidebarFocused)
			assert.Empty(t, mock.updates)
		})
	}
}
