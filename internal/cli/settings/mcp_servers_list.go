package settings

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
)

// MCPServerItem represents an MCP server in the list.
type MCPServerItem struct {
	Name      string
	Transport string
	Enabled   bool
}

// MCPServersListModel manages the MCP server list UI.
type MCPServersListModel struct {
	Servers  []MCPServerItem
	Cursor   int
	Selected string // Name of selected server, or "NEW"
	Deleted  string // Name of server to delete
	Exit     bool   // True if user wants to go back
}

// NewMCPServersListModel creates a new model from config.
func NewMCPServersListModel(cfg *config.Config) MCPServersListModel {
	var items []MCPServerItem
	if cfg.MCP.Servers != nil {
		for name, srv := range cfg.MCP.Servers {
			transport := srv.Transport
			if transport == "" {
				transport = "stdio"
			}
			items = append(items, MCPServerItem{
				Name:      name,
				Transport: transport,
				Enabled:   srv.IsEnabled(),
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return MCPServersListModel{
		Servers: items,
		Cursor:  0,
	}
}

// Init implements tea.Model.
func (m MCPServersListModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for the MCP server list.
func (m MCPServersListModel) Update(msg tea.Msg) (MCPServersListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Servers) {
				m.Cursor++
			}
		case "enter":
			if m.Cursor == len(m.Servers) {
				m.Selected = "NEW"
			} else {
				m.Selected = m.Servers[m.Cursor].Name
			}
			return m, nil
		case "d":
			if m.Cursor < len(m.Servers) {
				m.Deleted = m.Servers[m.Cursor].Name
				return m, nil
			}
		case "esc":
			m.Exit = true
			return m, nil
		}
	}
	return m, nil
}

// View renders the MCP server list.
func (m MCPServersListModel) View() string {
	var b strings.Builder

	// Items inside a container
	var body strings.Builder
	for i, srv := range m.Servers {
		cursor := "  "
		itemStyle := lipgloss.NewStyle()

		if m.Cursor == i {
			cursor = tui.CursorStyle.Render("▸ ")
			itemStyle = tui.ActiveItemStyle
		}

		body.WriteString(cursor)
		status := "enabled"
		if !srv.Enabled {
			status = "disabled"
		}
		label := fmt.Sprintf("%s (%s) [%s]", srv.Name, srv.Transport, status)
		body.WriteString(itemStyle.Render(label))
		body.WriteString("\n")
	}

	// "Add New" item
	cursor := "  "
	var itemStyle lipgloss.Style
	if m.Cursor == len(m.Servers) {
		cursor = tui.CursorStyle.Render("▸ ")
		itemStyle = tui.ActiveItemStyle
	} else {
		itemStyle = lipgloss.NewStyle().Foreground(tui.Muted)
	}
	body.WriteString(cursor)
	body.WriteString(itemStyle.Render("+ Add New MCP Server"))

	// Wrap in container
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Muted).
		Padding(1, 2)
	b.WriteString(container.Render(body.String()))

	// Help footer
	b.WriteString("\n")
	b.WriteString(tui.HelpBar(
		tui.HelpEntry("↑↓", "Navigate"),
		tui.HelpEntry("Enter", "Select"),
		tui.HelpEntry("d", "Delete"),
		tui.HelpEntry("Esc", "Back"),
	))

	return b.String()
}
