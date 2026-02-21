package settings

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/langowarny/lango/internal/config"
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

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Manage Providers"))
	b.WriteString("\n\n")

	for i, p := range m.Providers {
		cursor := "  "
		itemStyle := lipgloss.NewStyle()

		if m.Cursor == i {
			cursor = "\u25b8 "
			itemStyle = itemStyle.Foreground(lipgloss.Color("#04B575")).Bold(true)
		}

		b.WriteString(cursor)
		label := fmt.Sprintf("%s (%s)", p.ID, p.Type)
		b.WriteString(itemStyle.Render(label))
		b.WriteString("\n")
	}

	cursor := "  "
	itemStyle := lipgloss.NewStyle()
	if m.Cursor == len(m.Providers) {
		cursor = "\u25b8 "
		itemStyle = itemStyle.Foreground(lipgloss.Color("#04B575")).Bold(true)
	}
	b.WriteString(cursor)
	b.WriteString(itemStyle.Render("+ Add New Provider"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("\u2191/\u2193: navigate \u2022 enter: select \u2022 d: delete \u2022 esc: back"))

	return b.String()
}
