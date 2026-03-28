// Package pages implements the individual pages for the Lango Cockpit TUI.
package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/cockpit/theme"
	"github.com/langoai/lango/internal/toolcatalog"
)

// toolsKeyMap holds the key bindings for the tools page.
type toolsKeyMap struct {
	Up   key.Binding
	Down key.Binding
	Back key.Binding
}

func defaultToolsKeyMap() toolsKeyMap {
	return toolsKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "navigate up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "navigate down"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// ToolsPage is a read-only catalog browser that lists tool categories
// on the left and tool details for the selected category on the right.
type ToolsPage struct {
	catalog        *toolcatalog.Catalog
	categories     []toolcatalog.Category
	tools          []toolcatalog.ToolSchema
	categoryCursor int
	keymap         toolsKeyMap
	width, height  int
}

// NewToolsPage creates a ToolsPage backed by the given catalog.
func NewToolsPage(catalog *toolcatalog.Catalog) *ToolsPage {
	p := &ToolsPage{
		catalog: catalog,
		keymap:  defaultToolsKeyMap(),
	}
	p.refreshCategories()
	return p
}

// refreshCategories reloads categories and tools from the catalog.
func (p *ToolsPage) refreshCategories() {
	p.categories = p.catalog.ListCategories()
	if p.categoryCursor >= len(p.categories) {
		p.categoryCursor = max(0, len(p.categories)-1)
	}
	p.refreshTools()
}

// refreshTools reloads the tool list for the currently selected category.
func (p *ToolsPage) refreshTools() {
	if len(p.categories) == 0 {
		p.tools = nil
		return
	}
	cat := p.categories[p.categoryCursor]
	p.tools = p.catalog.ListTools(cat.Name)
}

// --- tea.Model implementation ---

// Init implements tea.Model.
func (p *ToolsPage) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (p *ToolsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keymap.Up):
			if p.categoryCursor > 0 {
				p.categoryCursor--
				p.refreshTools()
			}
		case key.Matches(msg, p.keymap.Down):
			if p.categoryCursor < len(p.categories)-1 {
				p.categoryCursor++
				p.refreshTools()
			}
		}
	}
	return p, nil
}

// View implements tea.Model.
func (p *ToolsPage) View() string {
	if p.width == 0 {
		return ""
	}

	leftWidth := max(p.width*3/10, 16)
	rightWidth := p.width - leftWidth - 1 // 1 for separator

	left := p.renderCategories(leftWidth)
	right := p.renderToolDetails(rightWidth)

	sep := borderStyle.Render(strings.Repeat("│\n", max(p.height, 1)))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, sep, right)
}

// --- Page lifecycle ---

// Title implements cockpit.Page.
func (p *ToolsPage) Title() string { return "Tools" }

// ShortHelp implements cockpit.Page.
func (p *ToolsPage) ShortHelp() []key.Binding {
	return []key.Binding{p.keymap.Up, p.keymap.Down, p.keymap.Back}
}

// Activate implements cockpit.Page.
func (p *ToolsPage) Activate() tea.Cmd {
	p.refreshCategories()
	return nil
}

// Deactivate implements cockpit.Page.
func (p *ToolsPage) Deactivate() {}

// --- rendering helpers ---

// Reusable styles — allocated once, not per-render.
var (
	cursorStyle      = lipgloss.NewStyle().Foreground(theme.Primary)
	activeNameStyle  = lipgloss.NewStyle().Foreground(theme.TextPrimary).Bold(true)
	activeCountStyle = lipgloss.NewStyle().Foreground(theme.TextSecondary)
	mutedStyle       = lipgloss.NewStyle().Foreground(theme.Muted)
	headerStyle      = lipgloss.NewStyle().Foreground(theme.TextSecondary).Bold(true)
	borderStyle      = lipgloss.NewStyle().Foreground(theme.BorderDefault)
)

// categoryLine renders a single category row with cursor, name, and tool count.
func categoryLine(cat toolcatalog.Category, active bool, width int, toolCount int) string {
	countLabel := fmt.Sprintf("[%d tools]", toolCount)

	if active {
		cursor := cursorStyle.Render("  ▸ ")
		name := activeNameStyle.Render(cat.Name)
		count := activeCountStyle.Render(countLabel)
		return cursor + name + padding(width, 4+len(cat.Name)+len(countLabel)) + count
	}

	name := mutedStyle.Render(cat.Name)
	count := mutedStyle.Render(countLabel)
	return "    " + name + padding(width, 4+len(cat.Name)+len(countLabel)) + count
}

// padding returns spaces to fill the remaining width.
func padding(totalWidth, usedWidth int) string {
	n := totalWidth - usedWidth
	if n < 1 {
		return " "
	}
	return strings.Repeat(" ", n)
}

// renderCategories renders the left-side category list.
func (p *ToolsPage) renderCategories(width int) string {
	header := headerStyle.Width(width).Render("  CATEGORIES")

	lines := make([]string, 0, len(p.categories)+2)
	lines = append(lines, header, "")

	for i, cat := range p.categories {
		toolCount := len(p.catalog.ToolNamesForCategory(cat.Name))
		active := i == p.categoryCursor
		lines = append(lines, categoryLine(cat, active, width, toolCount))
	}

	// Fill remaining height.
	for len(lines) < p.height {
		lines = append(lines, strings.Repeat(" ", width))
	}

	return lipgloss.NewStyle().
		Width(width).
		Render(strings.Join(lines, "\n"))
}

// renderToolDetails renders the right-side tool table for the selected
// category.
func (p *ToolsPage) renderToolDetails(width int) string {
	if len(p.categories) == 0 {
		return mutedStyle.Width(width).Render("  No categories registered.")
	}

	cat := p.categories[p.categoryCursor]

	header := headerStyle.Width(width).
		Render(fmt.Sprintf("  %s — %s", strings.ToUpper(cat.Name), cat.Description))

	if len(p.tools) == 0 {
		return header + "\n\n" + mutedStyle.Width(width).
			Render("  No tools in this category.")
	}

	// Column widths: name 20, description flexible, safety 10.
	nameCol := 20
	safetyCol := 10
	descCol := width - nameCol - safetyCol - 6 // 6 for padding/separators
	if descCol < 10 {
		descCol = 10
	}

	// Table header row.
	tableHeader := fmt.Sprintf("  %-*s %-*s %-*s",
		nameCol, headerStyle.Render("Name"),
		descCol, headerStyle.Render("Description"),
		safetyCol, headerStyle.Render("Safety"),
	)
	divider := borderStyle.
		Render("  " + strings.Repeat("─", min(width-4, nameCol+descCol+safetyCol+4)))

	lines := make([]string, 0, len(p.tools)+4)
	lines = append(lines, header, "")
	lines = append(lines, tableHeader, divider)

	for _, t := range p.tools {
		name := truncate(t.Name, nameCol)
		desc := truncate(t.Description, descCol)
		safety := safetyStyle(t.SafetyLevel).Render(truncate(t.SafetyLevel, safetyCol))

		line := fmt.Sprintf("  %-*s %-*s %s", nameCol, name, descCol, desc, safety)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// truncate shortens s to maxLen, appending "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Pre-allocated safety level styles.
var safetyStyles = map[string]lipgloss.Style{
	"safe":      lipgloss.NewStyle().Foreground(theme.Success),
	"moderate":  lipgloss.NewStyle().Foreground(theme.Warning),
	"dangerous": lipgloss.NewStyle().Foreground(theme.Error),
}

// safetyStyle returns a styled renderer based on the safety level string.
func safetyStyle(level string) lipgloss.Style {
	if s, ok := safetyStyles[level]; ok {
		return s
	}
	return mutedStyle
}
