package sidebar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// PageSelectedMsg is sent when the user selects a sidebar item.
type PageSelectedMsg struct {
	ID string
}

// MenuItem represents a single navigation entry in the sidebar.
type MenuItem struct {
	ID       string
	Icon     string
	Label    string
	Disabled bool
}

// Model is the Bubble Tea model for the cockpit sidebar.
type Model struct {
	items   []MenuItem
	active  string
	cursor  int
	focused bool
	visible bool
	height  int
}

// New creates a sidebar with the default menu items.
// Chat is active by default; the sidebar is visible by default.
func New() Model {
	return Model{
		items: []MenuItem{
			{ID: "chat", Icon: theme.IconChat, Label: "Chat"},
			{ID: "settings", Icon: theme.IconSettings, Label: "Settings"},
			{ID: "tools", Icon: theme.IconTools, Label: "Tools"},
			{ID: "status", Icon: theme.IconStatus, Label: "Status"},
			{ID: "sessions", Icon: theme.IconSessions, Label: "Sessions"},
		},
		active:  "chat",
		visible: true,
	}
}

// Init satisfies tea.Model. No initial command is needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model. Handles mouse events unconditionally;
// keyboard events require focus.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		// Mouse clicks work regardless of focused state.
		if msg.Action == tea.MouseActionRelease {
			idx := msg.Y
			if idx >= 0 && idx < len(m.items) && !m.items[idx].Disabled {
				return m, func() tea.Msg {
					return PageSelectedMsg{ID: m.items[idx].ID}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			m.moveCursorUp()
		case "down", "j":
			m.moveCursorDown()
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.items) && !m.items[m.cursor].Disabled {
				id := m.items[m.cursor].ID
				return m, func() tea.Msg { return PageSelectedMsg{ID: id} }
			}
		}
	}
	return m, nil
}

func (m *Model) moveCursorUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.items[i].Disabled {
			m.cursor = i
			return
		}
	}
}

func (m *Model) moveCursorDown() {
	for i := m.cursor + 1; i < len(m.items); i++ {
		if !m.items[i].Disabled {
			m.cursor = i
			return
		}
	}
}

// Sidebar styles — created once, reused across View() calls.
var (
	contentWidth = theme.SidebarFullWidth - 1

	activeIconStyle = lipgloss.NewStyle().
			Foreground(theme.Primary)

	activeLabelStyle = lipgloss.NewStyle().
				Foreground(theme.TextPrimary).
				Bold(true)

	inactiveStyle = lipgloss.NewStyle().
			Foreground(theme.Muted)

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4B5563"))

	accentBarStyle = lipgloss.NewStyle().
			Foreground(theme.Accent)

	sidebarCursorStyle = lipgloss.NewStyle().
				Foreground(theme.Primary)

	rowStyle = lipgloss.NewStyle().
			Width(contentWidth).
			Background(theme.Surface0)

	borderStyleSB = lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRightForeground(theme.BorderDefault).
			Width(theme.SidebarFullWidth).
			Background(theme.Surface0)
)

// View renders the sidebar. Returns "" when not visible.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	capacity := len(m.items)
	if m.height > capacity {
		capacity = m.height
	}
	rows := make([]string, 0, capacity)

	for i, it := range m.items {
		var line string
		switch {
		case it.Disabled:
			line = " " + disabledStyle.Render(it.Icon+" "+it.Label)
		case it.ID == m.active:
			bar := accentBarStyle.Render("┃")
			icon := activeIconStyle.Render(it.Icon)
			label := activeLabelStyle.Render(it.Label)
			if m.focused && i == m.cursor {
				line = sidebarCursorStyle.Render("▸") + icon + " " + label
			} else {
				line = bar + icon + " " + label
			}
		default:
			if m.focused && i == m.cursor {
				line = sidebarCursorStyle.Render("▸") + inactiveStyle.Render(it.Icon+" "+it.Label)
			} else {
				line = " " + inactiveStyle.Render(it.Icon+" "+it.Label)
			}
		}
		rows = append(rows, rowStyle.Render(line))
	}

	// Fill remaining height with empty lines.
	if m.height > len(rows) {
		filler := rowStyle.Render("")
		for i := len(rows); i < m.height; i++ {
			rows = append(rows, filler)
		}
	}

	return borderStyleSB.Render(strings.Join(rows, "\n"))
}

// SetHeight stores the available height so View can fill to that size.
func (m *Model) SetHeight(h int) {
	m.height = h
}

// SetVisible shows or hides the sidebar.
func (m *Model) SetVisible(v bool) {
	m.visible = v
}

// SetActive changes the highlighted menu item by ID.
func (m *Model) SetActive(id string) {
	m.active = id
}

// SetFocused controls whether the sidebar accepts key input.
func (m *Model) SetFocused(f bool) {
	m.focused = f
}
