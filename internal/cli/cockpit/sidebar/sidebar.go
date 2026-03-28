package sidebar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
)

// MenuItem represents a single navigation entry in the sidebar.
type MenuItem struct {
	ID    string
	Icon  string
	Label string
}

// Model is the Bubble Tea model for the cockpit sidebar.
type Model struct {
	items   []MenuItem
	active  string
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

// Update satisfies tea.Model. The sidebar is non-interactive in Change-1,
// so it returns itself unchanged for every message.
func (m Model) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// Sidebar styles — created once, reused across View() calls.
var (
	contentWidth = theme.SidebarFullWidth - 1 // reserve 1 col for right border

	activeIconStyle = lipgloss.NewStyle().
			Foreground(theme.Primary)

	activeLabelStyle = lipgloss.NewStyle().
				Foreground(theme.TextPrimary).
				Bold(true)

	inactiveStyle = lipgloss.NewStyle().
			Foreground(theme.Muted)

	accentBarStyle = lipgloss.NewStyle().
			Foreground(theme.Accent)

	rowStyle = lipgloss.NewStyle().
			Width(contentWidth).
			Background(theme.Surface0)

	borderStyle = lipgloss.NewStyle().
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

	for _, it := range m.items {
		var line string
		if it.ID == m.active {
			bar := accentBarStyle.Render("┃")
			icon := activeIconStyle.Render(it.Icon)
			label := activeLabelStyle.Render(it.Label)
			line = bar + icon + " " + label
		} else {
			line = " " + inactiveStyle.Render(it.Icon+" "+it.Label)
		}
		// Pad to contentWidth so the background fills evenly.
		rows = append(rows, rowStyle.Render(line))
	}

	// Fill remaining height with empty lines.
	if m.height > len(rows) {
		filler := rowStyle.Render("")
		for i := len(rows); i < m.height; i++ {
			rows = append(rows, filler)
		}
	}

	return borderStyle.Render(strings.Join(rows, "\n"))
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
