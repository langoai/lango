package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
)

// renderApprovalStrip renders a Tier 1 compact single-line approval strip.
func renderApprovalStrip(vm approval.ApprovalViewModel, state *approvalState, width int) string {
	isConfirmPending := state != nil && state.confirmPending
	confirmKey := "a"
	if state != nil && strings.TrimSpace(state.confirmAction) != "" {
		confirmKey = state.confirmAction
	}
	stripWidth := max(width, 1)
	safeToolName := sanitizeDisplayText(vm.Request.ToolName)

	toolBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Warning).
		Render(safeToolName)

	if vm.Risk.Level == "critical" {
		toolBadge += " " + lipgloss.NewStyle().
			Foreground(tui.Error).
			Bold(true).
			Render("(destructive)")
	}

	summary := vm.Request.Summary
	if summary == "" {
		summary = fmt.Sprintf("Execute tool: %s", safeToolName)
	}
	summary = sanitizeDisplayText(summary)

	if badge := formatChannelBadge(vm.Request.SessionKey); badge != "" {
		summary = badge + " " + summary
	}

	var keys string
	if isConfirmPending {
		keys = lipgloss.NewStyle().
			Foreground(tui.Warning).
			Bold(true).
			Render(fmt.Sprintf("Press '%s' again to confirm  d/Esc denies", confirmKey))
	} else {
		keys = lipgloss.NewStyle().
			Foreground(tui.Muted).
			Render("[a]llow  [s] allow session  [d/Esc] deny")
	}

	// Compute available space for summary by subtracting fixed elements.
	// The content layout is " %s  %s  %s" which adds 6 characters of spacing.
	usedWidth := lipgloss.Width(toolBadge) + lipgloss.Width(keys) + 6
	maxSummary := max(stripWidth-usedWidth, 0)

	if maxSummary == 0 {
		summary = ""
	} else {
		summary = ansi.Truncate(summary, maxSummary, "…")
	}

	summaryText := lipgloss.NewStyle().
		Foreground(tui.Foreground).
		Render(summary)

	content := fmt.Sprintf(" %s  %s  %s", toolBadge, summaryText, keys)

	// Truncate the assembled content to guarantee single-line output on very narrow terminals.
	content = ansi.Truncate(content, stripWidth, "")

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(tui.Foreground).
		Width(stripWidth).
		MaxHeight(1).
		Render(content)
}
