package cockpit

import (
	"testing"

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

func (m *mockChild) Init() tea.Cmd { return nil }

func (m *mockChild) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updates = append(m.updates, msg)
	return m, nil
}

func (m *mockChild) View() string { return m.viewContent }

func (m *mockChild) SetProgram(_ *tea.Program) { m.programSet = true }

// newTestModel creates a cockpit Model with a mock child.
func newTestModel(mock *mockChild) *Model {
	return &Model{
		child:          mock,
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
	m.width = 120
	m.height = 40

	m.Update(ctrlB())

	// After Ctrl+B, sidebar is hidden, so mock should receive a
	// WindowSizeMsg with full width (120).
	require.Len(t, mock.updates, 1)
	wsm, ok := mock.updates[0].(tea.WindowSizeMsg)
	require.True(t, ok, "expected WindowSizeMsg, got %T", mock.updates[0])
	assert.Equal(t, 120, wsm.Width)
	assert.Equal(t, 40, wsm.Height)
}

func TestCtrlB_WidthCalculation(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.width = 120
	m.height = 40
	m.sidebarVisible = true

	// Initial WindowSizeMsg with sidebar visible.
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	require.Len(t, mock.updates, 1)

	wsm1, ok := mock.updates[0].(tea.WindowSizeMsg)
	require.True(t, ok)
	assert.Equal(t, 120-theme.SidebarFullWidth, wsm1.Width,
		"child should receive width minus sidebar width")

	// Toggle sidebar off.
	m.Update(ctrlB())
	require.Len(t, mock.updates, 2)

	wsm2, ok := mock.updates[1].(tea.WindowSizeMsg)
	require.True(t, ok)
	assert.Equal(t, 120, wsm2.Width,
		"child should receive full width when sidebar is hidden")
}

func TestWindowSizeMsg_ReducedWidth(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.sidebarVisible = true

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	require.Len(t, mock.updates, 1)
	wsm, ok := mock.updates[0].(tea.WindowSizeMsg)
	require.True(t, ok)
	assert.Equal(t, 100, wsm.Width, "120 - 20 sidebar = 100")
	assert.Equal(t, 40, wsm.Height)
}

func TestCockpitOnly_CtrlB(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)
	m.width = 120
	m.height = 40

	m.Update(ctrlB())

	// Ctrl+B should NOT be forwarded as a key message; the child should
	// only see the synthetic WindowSizeMsg.
	for _, msg := range mock.updates {
		_, isKey := msg.(tea.KeyMsg)
		assert.False(t, isKey, "Ctrl+B key should not be forwarded to child")
	}
}

func TestSetProgram_Delegation(t *testing.T) {
	mock := &mockChild{}
	m := newTestModel(mock)

	m.SetProgram(nil)

	assert.True(t, mock.programSet, "SetProgram should delegate to child")
}
