package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// DependencyPanel shows prerequisite status for a category.
// It is a pure rendering component with cursor navigation.
type DependencyPanel struct {
	Results    []DepResult
	Cursor     int
	CategoryID string
}

// NewDependencyPanel creates a panel for the given category's dependency results.
// Returns nil if all dependencies are met (no panel needed).
func NewDependencyPanel(categoryID string, results []DepResult) *DependencyPanel {
	if len(results) == 0 {
		return nil
	}
	// Check if any dependency is unmet
	hasUnmet := false
	for _, r := range results {
		if r.Status != DepMet {
			hasUnmet = true
			break
		}
	}
	if !hasUnmet {
		return nil
	}
	return &DependencyPanel{
		Results:    results,
		Cursor:     0,
		CategoryID: categoryID,
	}
}

// View renders the dependency panel.
func (p *DependencyPanel) View() string {
	var b strings.Builder

	// Panel header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Warning)
	b.WriteString(headerStyle.Render("Prerequisites"))
	b.WriteString("\n")

	for i, r := range p.Results {
		isSelected := i == p.Cursor
		prefix := "  "
		if isSelected {
			prefix = tui.CursorStyle.Render("▸ ")
		}

		var indicator string
		var labelStyle lipgloss.Style

		switch r.Status {
		case DepMet:
			indicator = tui.SuccessStyle.Render(tui.CheckPass)
			labelStyle = lipgloss.NewStyle().Foreground(tui.Success)
		case DepNotEnabled:
			indicator = tui.ErrorStyle.Render(tui.CheckFail)
			labelStyle = lipgloss.NewStyle().Foreground(tui.Error)
		case DepMisconfigured:
			indicator = tui.WarningStyle.Render(tui.CheckWarn)
			labelStyle = lipgloss.NewStyle().Foreground(tui.Warning)
		}

		if isSelected {
			labelStyle = labelStyle.Bold(true)
		}

		requiredTag := ""
		if !r.Required {
			requiredTag = lipgloss.NewStyle().Foreground(tui.Dim).Render(" (optional)")
		}

		b.WriteString(prefix)
		b.WriteString(indicator)
		b.WriteString(" ")
		b.WriteString(labelStyle.Render(r.Label))
		b.WriteString(requiredTag)

		// Show fix hint for unmet dependencies on selected line
		if isSelected && r.Status != DepMet {
			b.WriteString("\n")
			hintStyle := lipgloss.NewStyle().Foreground(tui.Dim).Italic(true).PaddingLeft(5)
			b.WriteString(hintStyle.Render(r.FixHint))
		}

		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	help := tui.HelpBar(
		tui.HelpEntry("↑↓", "Navigate"),
		tui.HelpEntry("Enter", "Jump to setting"),
		tui.HelpEntry("s", "Guided setup"),
		tui.HelpEntry("Tab", "Skip to form"),
	)
	b.WriteString(help)

	// Wrap in bordered box
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Warning).
		Padding(0, 1).
		MarginBottom(1)

	return container.Render(b.String())
}

// SelectedCategoryID returns the category ID at the current cursor position.
func (p *DependencyPanel) SelectedCategoryID() string {
	if p.Cursor >= 0 && p.Cursor < len(p.Results) {
		return p.Results[p.Cursor].CategoryID
	}
	return ""
}

// MoveUp moves the cursor up.
func (p *DependencyPanel) MoveUp() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

// MoveDown moves the cursor down.
func (p *DependencyPanel) MoveDown() {
	if p.Cursor < len(p.Results)-1 {
		p.Cursor++
	}
}

// UnmetCount returns the number of unmet (required) dependencies.
func (p *DependencyPanel) UnmetCount() int {
	count := 0
	for _, r := range p.Results {
		if r.Required && r.Status != DepMet {
			count++
		}
	}
	return count
}

// SelectedIsUnmet returns true if the currently selected dependency is unmet.
func (p *DependencyPanel) SelectedIsUnmet() bool {
	if p.Cursor >= 0 && p.Cursor < len(p.Results) {
		return p.Results[p.Cursor].Status != DepMet
	}
	return false
}

// StatusSummary returns a short summary like "1/3 met".
func (p *DependencyPanel) StatusSummary() string {
	met, total := 0, len(p.Results)
	for _, r := range p.Results {
		if r.Status == DepMet {
			met++
		}
	}
	return fmt.Sprintf("%d/%d met", met, total)
}
