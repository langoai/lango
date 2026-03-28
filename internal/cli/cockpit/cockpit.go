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
	pages          map[PageID]Page
	activePage     PageID
	sidebar        sidebar.Model
	keymap         keyMap
	sidebarVisible bool
	sidebarFocused bool
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
		pages:          make(map[PageID]Page),
		activePage:     PageChat,
		sidebar:        sidebar.New(),
		keymap:         defaultKeyMap(),
		sidebarVisible: true,
	}
}

// RegisterPage adds a page to the cockpit.
func (m *Model) RegisterPage(id PageID, page Page) {
	m.pages[id] = page
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
		childSize := tea.WindowSizeMsg{
			Width:  msg.Width - sw,
			Height: msg.Height,
		}
		// Forward to chat child.
		updated, cmd := m.child.Update(childSize)
		m.child = updated.(childModel)
		cmds := []tea.Cmd{cmd}
		// Forward to all registered pages.
		for id, page := range m.pages {
			up, c := page.Update(childSize)
			m.pages[id] = up.(Page)
			if c != nil {
				cmds = append(cmds, c)
			}
		}
		return m, tea.Batch(cmds...)

	case sidebar.PageSelectedMsg:
		target := PageIDFromString(msg.ID)
		cmd := m.switchPage(target)
		return m, cmd

	case tea.KeyMsg:
		// Global keys — always consumed regardless of focus.
		switch {
		case key.Matches(msg, m.keymap.ToggleSidebar):
			m.sidebarVisible = !m.sidebarVisible
			sw := m.sidebarWidth()
			updated, cmd := m.child.Update(tea.WindowSizeMsg{
				Width: m.width - sw, Height: m.height,
			})
			m.child = updated.(childModel)
			return m, cmd
		case key.Matches(msg, m.keymap.FocusToggle):
			m.sidebarFocused = !m.sidebarFocused
			m.sidebar.SetFocused(m.sidebarFocused)
			return m, nil
		case key.Matches(msg, m.keymap.Page1):
			return m, m.switchPage(PageChat)
		case key.Matches(msg, m.keymap.Page2):
			return m, m.switchPage(PageSettings)
		case key.Matches(msg, m.keymap.Page3):
			return m, m.switchPage(PageTools)
		case key.Matches(msg, m.keymap.Page4):
			return m, m.switchPage(PageStatus)
		}

		// Focus-dependent routing.
		if m.sidebarFocused {
			up, cmd := m.sidebar.Update(msg)
			m.sidebar = up.(sidebar.Model)
			return m, cmd
		}
	}

	// consume-or-forward to active page.
	if m.activePage == PageChat {
		updated, cmd := m.child.Update(msg)
		m.child = updated.(childModel)
		return m, cmd
	}
	if page, ok := m.pages[m.activePage]; ok {
		up, cmd := page.Update(msg)
		m.pages[m.activePage] = up.(Page)
		return m, cmd
	}
	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	var mainView string
	if m.activePage == PageChat {
		mainView = m.child.View()
	} else if page, ok := m.pages[m.activePage]; ok {
		mainView = page.View()
	}
	if m.sidebarVisible {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.sidebar.View(),
			mainView,
		)
	}
	return mainView
}

func (m *Model) sidebarWidth() int {
	if !m.sidebarVisible {
		return 0
	}
	return theme.SidebarFullWidth
}

func (m *Model) switchPage(target PageID) tea.Cmd {
	if target == m.activePage {
		return nil
	}
	// Deactivate old page.
	if m.activePage != PageChat {
		if old, ok := m.pages[m.activePage]; ok {
			old.Deactivate()
		}
	}
	m.activePage = target
	m.sidebar.SetActive(target.String())
	m.sidebarFocused = false
	m.sidebar.SetFocused(false)
	// Activate new page.
	if target != PageChat {
		if page, ok := m.pages[target]; ok {
			return page.Activate()
		}
	}
	return nil
}
