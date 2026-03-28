package cockpit

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit/sidebar"
	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// childModel is the minimal contract for the wrapped child.
// ChatModel satisfies this. Tests can use a mock.
type childModel interface {
	tea.Model
	SetProgram(p *tea.Program)
}

// Compile-time interface check.
var _ childModel = (*chat.ChatModel)(nil)

// Model is the root cockpit tea.Model.
type Model struct {
	child          childModel
	sidebar        sidebar.Model
	keymap         keyMap
	sidebarVisible bool
	width          int
	height         int
}

// New creates a cockpit Model wrapping a ChatModel.
func New(deps Deps) *Model {
	chatModel := chat.New(chat.Deps{
		TurnRunner: deps.TurnRunner,
		Config:     deps.Config,
		SessionKey: deps.SessionKey,
	})

	return &Model{
		child:          chatModel,
		sidebar:        sidebar.New(),
		keymap:         defaultKeyMap(),
		sidebarVisible: true,
	}
}

// SetProgram delegates to the wrapped child.
func (m *Model) SetProgram(p *tea.Program) {
	m.child.SetProgram(p)
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return m.child.Init()
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		sw := m.sidebarWidth()
		m.sidebar.SetHeight(msg.Height)
		updated, cmd := m.child.Update(tea.WindowSizeMsg{
			Width:  msg.Width - sw,
			Height: msg.Height,
		})
		m.child = updated.(childModel)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keymap.ToggleSidebar) {
			m.sidebarVisible = !m.sidebarVisible
			sw := m.sidebarWidth()
			updated, cmd := m.child.Update(tea.WindowSizeMsg{
				Width:  m.width - sw,
				Height: m.height,
			})
			m.child = updated.(childModel)
			return m, cmd
		}
	}

	// consume-or-forward: all other messages go to child.
	updated, cmd := m.child.Update(msg)
	m.child = updated.(childModel)
	return m, cmd
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.sidebarVisible {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.sidebar.View(),
			m.child.View(),
		)
	}
	return m.child.View()
}

func (m *Model) sidebarWidth() int {
	if !m.sidebarVisible {
		return 0
	}
	return theme.SidebarFullWidth
}
