package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
)

// ToolItemState represents the lifecycle state of a tool invocation.
type ToolItemState string

const (
	toolStateRunning         ToolItemState = "running"
	toolStateSuccess         ToolItemState = "success"
	toolStateError           ToolItemState = "error"
	toolStateCanceled        ToolItemState = "canceled"
	toolStateAwaitingApproval ToolItemState = "awaiting_approval"
)

// renderToolBlock renders a tool transcript item with state-specific icon and styling.
func renderToolBlock(toolName string, state ToolItemState, duration, output string, width int) string {
	icon, color := toolStateVisual(state)

	label := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(fmt.Sprintf("%s %s", icon, toolName))

	var detail string
	switch state {
	case toolStateRunning:
		detail = lipgloss.NewStyle().Foreground(tui.Muted).Render("running...")
	case toolStateSuccess:
		detail = lipgloss.NewStyle().Foreground(tui.Success).Render(fmt.Sprintf("(%s)", duration))
	case toolStateError:
		detail = lipgloss.NewStyle().Foreground(tui.Error).Render(fmt.Sprintf("failed (%s)", duration))
	case toolStateCanceled:
		detail = lipgloss.NewStyle().Foreground(tui.Muted).Render("canceled")
	case toolStateAwaitingApproval:
		detail = lipgloss.NewStyle().Foreground(tui.Warning).Render("awaiting approval")
	}

	line := fmt.Sprintf(" %s  %s", label, detail)

	if output != "" && (state == toolStateSuccess || state == toolStateError) {
		output = strings.ReplaceAll(output, "\n", " ")
		maxOutput := width - 4
		if maxOutput < 20 {
			maxOutput = 20
		}
		if lipgloss.Width(output) > maxOutput {
			output = ansi.Truncate(output, maxOutput, "…")
		}
		outputLine := lipgloss.NewStyle().
			Foreground(tui.Muted).
			PaddingLeft(4).
			Render(output)
		line += "\n" + outputLine
	}

	return line
}

func toolStateVisual(state ToolItemState) (string, lipgloss.Color) {
	switch state {
	case toolStateRunning:
		return "\u2699", tui.Warning // ⚙
	case toolStateSuccess:
		return "\u2713", tui.Success // ✓
	case toolStateError:
		return "\u2717", tui.Error // ✗
	case toolStateCanceled:
		return "\u2298", tui.Muted // ⊘
	case toolStateAwaitingApproval:
		return "\U0001F512", tui.Warning // 🔒
	default:
		return "\u2699", tui.Muted // ⚙
	}
}
