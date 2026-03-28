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
		keymap:         defaultKeyMap(),
		sidebarVisible: true,
		width:          120,
		height:         40,
	}
}

func ctrlB() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyCtrlB}
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
