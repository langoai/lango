package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
)

// renderApprovalStrip renders a Tier 1 compact single-line approval strip.
func renderApprovalStrip(vm approval.ApprovalViewModel, width int) string {
	stripWidth := max(width, 1)

	toolBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Warning).
		Render(vm.Request.ToolName)

	summary := vm.Request.Summary
	if summary == "" {
		summary = fmt.Sprintf("Execute tool: %s", vm.Request.ToolName)
	}

	// Truncate summary to fit on one line with keys.
	maxSummary := stripWidth - lipgloss.Width(toolBadge) - 45
	if maxSummary < 10 {
		maxSummary = 10
	}
	if len(summary) > maxSummary {
		summary = summary[:maxSummary-3] + "..."
	}

	summaryText := lipgloss.NewStyle().
		Foreground(tui.Foreground).
		Render(summary)

	keys := lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render("[a]llow  [s]ession  [d]eny")

	content := fmt.Sprintf(" %s  %s  %s", toolBadge, summaryText, keys)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(tui.Foreground).
		Width(stripWidth).
		Render(content)
}
