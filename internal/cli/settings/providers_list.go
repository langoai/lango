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

// ProviderItem represents a provider in the list.
type ProviderItem struct {
	ID   string
	Type string
}

// ProvidersListModel manages the provider list UI.
type ProvidersListModel struct {
	Providers []ProviderItem
	Cursor    int
	Selected  string // ID of selected provider, or "NEW"
	Deleted   string // ID of provider to delete
	Exit      bool   // True if user wants to go back
}

// NewProvidersListModel creates a new model from config.
func NewProvidersListModel(cfg *config.Config) ProvidersListModel {
	var items []ProviderItem
	if cfg.Providers != nil {
		for id, p := range cfg.Providers {
			items = append(items, ProviderItem{ID: id, Type: string(p.Type)})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return ProvidersListModel{
		Providers: items,
		Cursor:    0,
	}
}

// Init implements tea.Model.
func (m ProvidersListModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for the provider list.
func (m ProvidersListModel) Update(msg tea.Msg) (ProvidersListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Providers) {
				m.Cursor++
			}
		case "enter":
			if m.Cursor == len(m.Providers) {
				m.Selected = "NEW"
			} else {
				m.Selected = m.Providers[m.Cursor].ID
			}
			return m, nil
		case "d":
			if m.Cursor < len(m.Providers) {
				m.Deleted = m.Providers[m.Cursor].ID
				return m, nil
			}
		case "esc":
			m.Exit = true
			return m, nil
		}
	}
	return m, nil
}

// View renders the provider list.
func (m ProvidersListModel) View() string {
	var b strings.Builder

	// Items inside a container
	var body strings.Builder
	for i, p := range m.Providers {
		cursor := "  "
		itemStyle := lipgloss.NewStyle()

		if m.Cursor == i {
			cursor = tui.CursorStyle.Render("▸ ")
			itemStyle = tui.ActiveItemStyle
		}

		body.WriteString(cursor)
		label := fmt.Sprintf("%s (%s)", p.ID, p.Type)
		body.WriteString(itemStyle.Render(label))
		body.WriteString("\n")
	}

	// "Add New" item
	cursor := "  "
	itemStyle := lipgloss.NewStyle()
	if m.Cursor == len(m.Providers) {
		cursor = tui.CursorStyle.Render("▸ ")
		itemStyle = tui.ActiveItemStyle
	} else {
		itemStyle = lipgloss.NewStyle().Foreground(tui.Muted)
	}
	body.WriteString(cursor)
	body.WriteString(itemStyle.Render("+ Add New Provider"))

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
