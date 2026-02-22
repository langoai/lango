package onboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/langoai/lango/internal/cli/tui"
)

// StepDef describes a wizard step for display purposes.
type StepDef struct {
	Name string
}

// WizardSteps is the ordered list of wizard step definitions.
var WizardSteps = []StepDef{
	{Name: "Provider Setup"},
	{Name: "Agent Config"},
	{Name: "Channel Setup"},
	{Name: "Security & Auth"},
	{Name: "Test Configuration"},
}

// renderProgress renders the step indicator and progress bar.
func renderProgress(current int, width int) string {
	var b strings.Builder

	total := len(WizardSteps)
	stepName := ""
	if current >= 0 && current < total {
		stepName = WizardSteps[current].Name
	}

	// Header: [Step N/5] ━━━━━━━━━━ StepName
	header := fmt.Sprintf("[Step %d/%d]", current+1, total)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary)
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Foreground)

	// Progress bar
	barWidth := width - len(header) - len(stepName) - 4 // spaces + padding
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 40 {
		barWidth = 40
	}

	filled := 0
	if total > 0 {
		filled = (current + 1) * barWidth / total
	}

	filledStyle := lipgloss.NewStyle().Foreground(tui.Success)
	mutedStyle := lipgloss.NewStyle().Foreground(tui.Muted)

	bar := filledStyle.Render(strings.Repeat("\u2501", filled)) + mutedStyle.Render(strings.Repeat("\u2501", barWidth-filled))

	b.WriteString(headerStyle.Render(header))
	b.WriteString(" ")
	b.WriteString(bar)
	b.WriteString(" ")
	b.WriteString(nameStyle.Render(stepName))
	b.WriteString("\n\n")

	return b.String()
}

// renderStepList renders the vertical step list with status indicators.
func renderStepList(current int) string {
	var b strings.Builder

	successStyle := lipgloss.NewStyle().Foreground(tui.Success)
	highlightStyle := lipgloss.NewStyle().Foreground(tui.Highlight).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(tui.Muted)

	for i, step := range WizardSteps {
		var indicator string
		var style lipgloss.Style

		switch {
		case i < current:
			indicator = tui.CheckPass
			style = successStyle
		case i == current:
			indicator = "\u25b8"
			style = highlightStyle
		default:
			indicator = "\u25cb"
			style = mutedStyle
		}

		b.WriteString("  ")
		b.WriteString(style.Render(fmt.Sprintf("%s %s", indicator, step.Name)))
		b.WriteString("\n")
	}

	// Divider
	b.WriteString(mutedStyle.Render(strings.Repeat("\u2500", 40)))
	b.WriteString("\n\n")

	return b.String()
}
