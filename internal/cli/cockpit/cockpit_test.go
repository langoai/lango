package cockpit

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit/sidebar"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/observability"
)

// mockChild implements childModel for testing without real ChatModel.
type mockChild struct {
	updates     []tea.Msg
	programSet  bool
	viewContent string
}

func (m *mockChild) Init() tea.Cmd { return nil }
func (m *mockChild) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updates = append(m.updates, msg)
	return m, nil
}
func (m *mockChild) View() string              { return m.viewContent }
func (m *mockChild) SetProgram(_ *tea.Program) { m.programSet = true }

// mockPage implements Page for testing page routing.
type mockPage struct {
	title       string
	activated   bool
	deactivated bool
	updates     []tea.Msg
	viewContent string
}

func (p *mockPage) Init() tea.Cmd { return nil }
func (p *mockPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p.updates = append(p.updates, msg)
	return p, nil
}
func (p *mockPage) View() string             { return p.viewContent }
func (p *mockPage) Title() string            { return p.title }
func (p *mockPage) ShortHelp() []key.Binding { return nil }
func (p *mockPage) Activate() tea.Cmd        { p.activated = true; return nil }
func (p *mockPage) Deactivate()              { p.deactivated = true }

func newTestModel(mock *mockChild) *Model {
	return &Model{
		child:          mock,
		pages:          make(map[PageID]Page),
		activePage:     PageChat,
		sidebar:        sidebar.New(AllPageMetas()),
		contextPanel:   NewContextPanel(nil),
		keymap:         defaultKeyMap(),
		sidebarVisible: true,
		width:          120,
		height:         40,
	}
}

func newTestModelWithCollector(mock *mockChild) *Model {
	collector := observability.NewCollector()
	return &Model{
		child:          mock,
		pages:          make(map[PageID]Page),
		activePage:     PageChat,
		sidebar:        sidebar.New(AllPageMetas()),
		contextPanel:   NewContextPanel(collector),
		keymap:         defaultKeyMap(),
		sidebarVisible: true,
		width:          120,
		height:         40,
	}
}

func ctrlB() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyCtrlB}
}

func ctrlP() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyCtrlP}
}

func ctrlY() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyCtrlY}
}

func TestConsumeOrForward_ChunkMsg(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	msg := chat.ChunkMsg{Chunk: "hello"}
	m.Update(msg)
	require.Len(t, mock.updates, 1)
	assert.Equal(t, msg, mock.updates[0])
}

func TestConsumeOrForward_DoneMsg(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	msg := chat.DoneMsg{}
	m.Update(msg)
	require.Len(t, mock.updates, 1)
	assert.Equal(t, msg, mock.updates[0])
}

func TestCtrlB_SyntheticResize(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.Update(ctrlB())
	require.Len(t, mock.updates, 1)
	wsm, ok := mock.updates[0].(tea.WindowSizeMsg)
	require.True(t, ok)
	assert.Equal(t, 120, wsm.Width)
	assert.Equal(t, 40, wsm.Height)
}

func TestCtrlB_WidthCalculation(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	require.GreaterOrEqual(t, len(mock.updates), 1)
	wsm1 := mock.updates[0].(tea.WindowSizeMsg)
	assert.Equal(t, 120-theme.SidebarFullWidth, wsm1.Width)

	m.Update(ctrlB())
	last := mock.updates[len(mock.updates)-1].(tea.WindowSizeMsg)
	assert.Equal(t, 120, last.Width)
}

func TestWindowSizeMsg_ReducedWidth(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.sidebarVisible = true
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	require.GreaterOrEqual(t, len(mock.updates), 1)
	wsm := mock.updates[0].(tea.WindowSizeMsg)
	assert.Equal(t, 100, wsm.Width)
	assert.Equal(t, 40, wsm.Height)
}

func TestCockpitOnly_CtrlB(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.Update(ctrlB())
	for _, msg := range mock.updates {
		_, isKey := msg.(tea.KeyMsg)
		assert.False(t, isKey, "Ctrl+B key should not be forwarded to child")
	}
}

func TestSetProgram_Delegation(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.SetProgram(nil)
	assert.True(t, mock.programSet)
}

// --- New Change-2 tests ---

func TestPageRouting_SwitchToTools(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)

	cmd := m.switchPage(PageTools)
	assert.Equal(t, PageTools, m.activePage)
	assert.True(t, toolsPage.activated)
	_ = cmd
}

func TestPageRouting_DeactivateOld(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	statusPage := &mockPage{title: "Status"}
	m.RegisterPage(PageTools, toolsPage)
	m.RegisterPage(PageStatus, statusPage)

	m.switchPage(PageTools)
	m.switchPage(PageStatus)

	assert.True(t, toolsPage.deactivated, "old page should be deactivated")
	assert.True(t, statusPage.activated, "new page should be activated")
}

func TestPageRouting_SamePageNoop(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	cmd := m.switchPage(PageChat)
	assert.Nil(t, cmd, "switching to same page should be no-op")
}

func TestFocusToggle(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)

	assert.False(t, m.sidebarFocused)
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, m.sidebarFocused)
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.False(t, m.sidebarFocused)
}

func TestSidebarFocused_KeysGoToSidebar(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.sidebarFocused = true
	m.sidebar.SetFocused(true)

	initialChildUpdates := len(mock.updates)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Child should NOT receive the down key.
	assert.Equal(t, initialChildUpdates, len(mock.updates),
		"down key should go to sidebar, not child")
}

func TestPageSelectedMsg_SwitchesPage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)

	m.Update(sidebar.PageSelectedMsg{ID: "tools"})
	assert.Equal(t, PageTools, m.activePage)
	assert.True(t, toolsPage.activated)
	assert.False(t, m.sidebarFocused, "focus should return to content")
}

func TestPageSelectedMsg_SwitchesToDeadLettersPage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	deadLettersPage := &mockPage{title: "Dead Letters"}
	m.RegisterPage(PageDeadLetters, deadLettersPage)

	m.Update(sidebar.PageSelectedMsg{ID: "dead-letters"})
	assert.Equal(t, PageDeadLetters, m.activePage)
	assert.True(t, deadLettersPage.activated)
	assert.False(t, m.sidebarFocused, "focus should return to content")
}

func TestViewDispatchesToActivePage(t *testing.T) {
	mock := &mockChild{viewContent: "chat-view"}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools", viewContent: "tools-view"}
	m.RegisterPage(PageTools, toolsPage)

	// PageChat shows child view.
	view := m.View()
	assert.Contains(t, view, "chat-view")

	// Switch to tools.
	m.switchPage(PageTools)
	view = m.View()
	assert.Contains(t, view, "tools-view")
}

func TestForwardToActivePage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	// Send a key — should go to toolsPage, not child.
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Len(t, mock.updates, 0, "child should not receive keys when tools page is active")
	assert.Len(t, toolsPage.updates, 1, "tools page should receive the key")
}

// --- Change-3 W2: context panel tests ---

// --- Phase 3: Runtime message routing tests ---

func TestRuntimeMsg_DelegationReachesChatFromNonChatPage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	msg := chat.DelegationMsg{From: "operator", To: "librarian", Reason: "search"}
	m.Update(msg)

	require.Len(t, mock.updates, 1, "DelegationMsg must reach chat child even when Tools page is active")
	assert.Equal(t, msg, mock.updates[0])
	assert.Empty(t, toolsPage.updates, "DelegationMsg should not go to tools page")
}

func TestRuntimeMsg_BudgetWarningReachesChatFromNonChatPage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	msg := chat.BudgetWarningMsg{Used: 12, Max: 15}
	m.Update(msg)

	require.Len(t, mock.updates, 1, "BudgetWarningMsg must reach chat child from non-chat page")
	assert.Equal(t, msg, mock.updates[0])
}

func TestRuntimeMsg_RecoveryReachesChatFromNonChatPage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	msg := chat.RecoveryMsg{CauseClass: "rate_limit", Action: "retry", Attempt: 1}
	m.Update(msg)

	require.Len(t, mock.updates, 1, "RecoveryMsg must reach chat child from non-chat page")
}

func TestApprovalRequestMsg_SwitchesToChatAndForwards(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	assert.Equal(t, PageTools, m.activePage)

	msg := chat.ApprovalRequestMsg{
		Request: approval.ApprovalRequest{ToolName: "exec"},
	}
	m.Update(msg)

	// Must switch to Chat page.
	assert.Equal(t, PageChat, m.activePage, "ApprovalRequestMsg should switch to Chat page")
	// Must reach chat child.
	require.Len(t, mock.updates, 1, "ApprovalRequestMsg must reach chat child")
	// Tools page should NOT receive the message.
	assert.Empty(t, toolsPage.updates)
}

func TestApprovalRequestMsg_AlreadyOnChat(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)

	assert.Equal(t, PageChat, m.activePage)

	msg := chat.ApprovalRequestMsg{
		Request: approval.ApprovalRequest{ToolName: "fs_write"},
	}
	m.Update(msg)

	// Should still be on Chat and forwarded.
	assert.Equal(t, PageChat, m.activePage)
	require.Len(t, mock.updates, 1)
}

func TestRuntimeMsg_DoneMsgFlushesTokensThenForwards(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)

	// Set up a RuntimeTracker with some accumulated tokens.
	tracker := NewRuntimeTracker(nil, nil, "sess-1")
	tracker.StartTurn()
	// Manually set token values (no bus, so inject directly).
	tracker.mu.Lock()
	tracker.turnTokens = tokenSnapshot{InputTokens: 10, OutputTokens: 20, TotalTokens: 30, CacheTokens: 5}
	tracker.mu.Unlock()
	m.SetRuntimeTracker(tracker)

	doneMsg := chat.DoneMsg{}
	m.Update(doneMsg)

	// Child should receive DoneMsg FIRST, then TurnTokenUsageMsg.
	require.Len(t, mock.updates, 2, "child should receive DoneMsg and TurnTokenUsageMsg")
	assert.IsType(t, chat.DoneMsg{}, mock.updates[0], "first message should be DoneMsg")
	tokenMsg, ok := mock.updates[1].(chat.TurnTokenUsageMsg)
	require.True(t, ok, "second message should be TurnTokenUsageMsg")
	assert.Equal(t, int64(30), tokenMsg.TotalTokens)

	// Tracker should be reset.
	snap := tracker.Snapshot()
	assert.False(t, snap.IsRunning, "turn should no longer be running after DoneMsg")
	assert.Equal(t, 0, snap.DelegationCount)
}

func TestRuntimeMsg_DoneMsgNoTokensSkipsSummary(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	tracker := NewRuntimeTracker(nil, nil, "sess-1")
	tracker.StartTurn()
	m.SetRuntimeTracker(tracker)

	m.Update(chat.DoneMsg{})

	// Only DoneMsg, no TurnTokenUsageMsg (zero tokens).
	require.Len(t, mock.updates, 1, "child should only receive DoneMsg when tokens are zero")
	assert.IsType(t, chat.DoneMsg{}, mock.updates[0])
}

func TestRuntimeMsg_DelegationTracksOrchReturn(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	tracker := NewRuntimeTracker(nil, nil, "sess-1")
	tracker.StartTurn()
	m.SetRuntimeTracker(tracker)

	// Outward delegation.
	m.Update(chat.DelegationMsg{From: "orchestrator", To: "operator"})
	assert.Equal(t, 1, tracker.Snapshot().DelegationCount)
	assert.Equal(t, "operator", tracker.Snapshot().ActiveAgent)

	// Return hop — counter should NOT increment, but active agent updates.
	m.Update(chat.DelegationMsg{From: "operator", To: "lango-orchestrator"})
	assert.Equal(t, 1, tracker.Snapshot().DelegationCount, "return hop should not increment counter")
	assert.Equal(t, "lango-orchestrator", tracker.Snapshot().ActiveAgent, "active agent should update to orchestrator")
}

func TestRuntimeMsg_StartTurnOnFirstContentEvent(t *testing.T) {
	tests := []struct {
		give string
		msg  tea.Msg
	}{
		{"ToolStartedMsg", chat.ToolStartedMsg{CallID: "c1", ToolName: "test"}},
		{"ThinkingStartedMsg", chat.ThinkingStartedMsg{AgentName: "agent"}},
		{"ChunkMsg", chat.ChunkMsg{Chunk: "hello"}},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			mock := &mockChild{}
			m := newTestModel(mock)
			tracker := NewRuntimeTracker(nil, nil, "sess-1")
			m.SetRuntimeTracker(tracker)

			assert.False(t, tracker.Snapshot().IsRunning, "should not be running before content event")

			m.Update(tt.msg)

			assert.True(t, tracker.Snapshot().IsRunning, "should be running after content event")
		})
	}
}

func TestCtrlP_TogglesContext(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)

	assert.False(t, m.contextVisible)

	m.Update(ctrlP())
	assert.True(t, m.contextVisible)
	assert.True(t, m.contextPanel.tickActive)

	m.Update(ctrlP())
	assert.False(t, m.contextVisible)
	assert.False(t, m.contextPanel.tickActive)
}

func TestCtrlP_SyntheticResize(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)
	m.width = 120
	m.height = 40

	// Toggle context on — child should get reduced width.
	m.Update(ctrlP())
	require.GreaterOrEqual(t, len(mock.updates), 1)
	last := mock.updates[len(mock.updates)-1].(tea.WindowSizeMsg)
	expectedWidth := 120 - theme.SidebarFullWidth - theme.ContextPanelWidth
	assert.Equal(t, expectedWidth, last.Width,
		"child width should subtract both sidebar and context panel")
}

func TestCtrlP_FirstToggle_ContextPanelGetsCorrectWidth(t *testing.T) {
	// Reproduce: initial WindowSizeMsg with context hidden (width=0),
	// then first Ctrl+P toggle should send correct width to panel.
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)

	// Step 1: Initial resize with context hidden.
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	// Context panel should have received Width=0 (since hidden).
	assert.Equal(t, 0, m.contextPanel.width,
		"context panel should have width=0 while hidden")

	// Step 2: First Ctrl+P toggle.
	m.Update(ctrlP())
	assert.True(t, m.contextVisible)
	assert.Equal(t, theme.ContextPanelWidth, m.contextPanel.width,
		"context panel should receive ContextPanelWidth after first toggle")
}

func TestWindowSizeMsg_ThreePanelLayout(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)
	m.contextVisible = true
	m.contextPanel.SetVisible(true)

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	require.GreaterOrEqual(t, len(mock.updates), 1)
	wsm := mock.updates[0].(tea.WindowSizeMsg)
	expectedWidth := 120 - theme.SidebarFullWidth - theme.ContextPanelWidth
	assert.Equal(t, expectedWidth, wsm.Width)
}

func TestView_ThreePanelLayout(t *testing.T) {
	mock := &mockChild{viewContent: "main-content"}
	m := newTestModel(mock)
	m.contextVisible = true
	m.contextPanel.SetVisible(true)
	m.contextPanel.SetHeight(10)

	view := m.View()
	assert.Contains(t, view, "main-content")
}

func TestView_SidebarHiddenContextVisible(t *testing.T) {
	mock := &mockChild{viewContent: "main-content"}
	m := newTestModel(mock)
	m.sidebarVisible = false
	m.contextVisible = true
	m.contextPanel.SetVisible(true)
	m.contextPanel.SetHeight(10)

	view := m.View()
	assert.Contains(t, view, "main-content")
}

func TestContextPanelWidth_Visible(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.contextVisible = true
	assert.Equal(t, theme.ContextPanelWidth, m.contextPanelWidth())
}

func TestContextPanelWidth_Hidden(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.contextVisible = false
	assert.Equal(t, 0, m.contextPanelWidth())
}

func TestMouseRouting_SidebarRegion(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.sidebarVisible = true

	// Click within sidebar width (X < 20).
	msg := tea.MouseMsg{
		X:      5,
		Y:      1,
		Action: tea.MouseActionRelease,
	}
	m.Update(msg)
	// Child should NOT have received the mouse event.
	assert.Len(t, mock.updates, 0,
		"mouse click in sidebar region should not reach child")
}

func TestMouseRouting_ContentRegion(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.sidebarVisible = true

	// Click outside sidebar width (X >= 20).
	msg := tea.MouseMsg{
		X:      25,
		Y:      5,
		Action: tea.MouseActionRelease,
	}
	m.Update(msg)
	require.Len(t, mock.updates, 1, "mouse click in content region should reach child")
}

func TestMouseRouting_ActivePage(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	msg := tea.MouseMsg{
		X:      25,
		Y:      5,
		Action: tea.MouseActionRelease,
	}
	m.Update(msg)
	assert.Len(t, mock.updates, 0, "mouse should not reach child when tools page is active")
	assert.Len(t, toolsPage.updates, 1, "mouse should reach active page")
}

func TestCtrlY_CopiesActiveView(t *testing.T) {
	mock := &mockChild{viewContent: "test-clipboard-content"}
	m := newTestModel(mock)

	// This test verifies the code path doesn't panic.
	// Actual clipboard write may fail in CI but should not error.
	m.Update(ctrlY())
}

func TestCtrlY_CopiesPageView(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	toolsPage := &mockPage{title: "Tools", viewContent: "tools-clipboard"}
	m.RegisterPage(PageTools, toolsPage)
	m.switchPage(PageTools)

	// Should not panic.
	m.Update(ctrlY())
}

func TestWindowSizeMsg_PropagatesContextPanelSize(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)
	m.contextVisible = true
	m.contextPanel.SetVisible(true)

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Equal(t, 40, m.contextPanel.height)
}

func TestCtrlP_ResizePropagatesAllPages(t *testing.T) {
	mock := &mockChild{}
	m := newTestModelWithCollector(mock)
	toolsPage := &mockPage{title: "Tools"}
	m.RegisterPage(PageTools, toolsPage)
	m.width = 120
	m.height = 40

	m.Update(ctrlP())

	// Tools page should have received a resize message.
	require.GreaterOrEqual(t, len(toolsPage.updates), 1)
	wsm, ok := toolsPage.updates[0].(tea.WindowSizeMsg)
	require.True(t, ok, "page should receive WindowSizeMsg")
	expectedWidth := 120 - theme.SidebarFullWidth - theme.ContextPanelWidth
	assert.Equal(t, expectedWidth, wsm.Width)
}

func TestSidebarClick_UnregisteredPage_NoOp(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	// Do NOT register PageTools — it remains unregistered.

	// Simulate sidebar click selecting "tools".
	m.Update(sidebar.PageSelectedMsg{ID: "tools"})

	// activePage should change (switchPage sets it), but no crash.
	assert.Equal(t, PageTools, m.activePage,
		"activePage should update even for unregistered page")

	// Child should NOT have received any messages (no Activate forwarded).
	assert.Empty(t, mock.updates,
		"no messages should reach child for an unregistered page switch")
}
