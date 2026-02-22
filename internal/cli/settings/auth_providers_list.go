package settings

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/langoai/lango/internal/config"
)

// AuthProviderItem represents an OIDC provider in the list.
type AuthProviderItem struct {
	ID        string
	IssuerURL string
}

// AuthProvidersListModel manages the OIDC provider list UI.
type AuthProvidersListModel struct {
	Providers []AuthProviderItem
	Cursor    int
	Selected  string // ID of selected provider, or "NEW"
	Deleted   string // ID of provider to delete
	Exit      bool   // True if user wants to go back
}

// NewAuthProvidersListModel creates a new model from config.
func NewAuthProvidersListModel(cfg *config.Config) AuthProvidersListModel {
	var items []AuthProviderItem
	if cfg.Auth.Providers != nil {
		for id, p := range cfg.Auth.Providers {
			items = append(items, AuthProviderItem{ID: id, IssuerURL: p.IssuerURL})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return AuthProvidersListModel{
		Providers: items,
		Cursor:    0,
	}
}

// Init implements tea.Model.
func (m AuthProvidersListModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for the OIDC provider list.
func (m AuthProvidersListModel) Update(msg tea.Msg) (AuthProvidersListModel, tea.Cmd) {
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

// View renders the OIDC provider list.
func (m AuthProvidersListModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Manage OIDC Providers"))
	b.WriteString("\n\n")

	for i, p := range m.Providers {
		cursor := "  "
		itemStyle := lipgloss.NewStyle()

		if m.Cursor == i {
			cursor = "\u25b8 "
			itemStyle = itemStyle.Foreground(lipgloss.Color("#04B575")).Bold(true)
		}

		b.WriteString(cursor)
		label := fmt.Sprintf("%s (%s)", p.ID, p.IssuerURL)
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
	b.WriteString(itemStyle.Render("+ Add New OIDC Provider"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("\u2191/\u2193: navigate \u2022 enter: select \u2022 d: delete \u2022 esc: back"))

	return b.String()
}
