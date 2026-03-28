package cockpit

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func (m *mockChild) Init() tea.Cmd                           { return nil }
func (m *mockChild) Update(msg tea.Msg) (tea.Model, tea.Cmd) { m.updates = append(m.updates, msg); return m, nil }
func (m *mockChild) View() string                            { return m.viewContent }
func (m *mockChild) SetProgram(_ *tea.Program)               { m.programSet = true }

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
func (p *mockPage) View() string              { return p.viewContent }
func (p *mockPage) Title() string             { return p.title }
func (p *mockPage) ShortHelp() []key.Binding  { return nil }
func (p *mockPage) Activate() tea.Cmd         { p.activated = true; return nil }
func (p *mockPage) Deactivate()               { p.deactivated = true }

func newTestModel(mock *mockChild) *Model {
	return &Model{
		child:          mock,
		pages:          make(map[PageID]Page),
		activePage:     PageChat,
		sidebar:        sidebar.New(),
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
		sidebar:        sidebar.New(),
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
