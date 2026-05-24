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
	toolStateRunning          ToolItemState = "running"
	toolStateSuccess          ToolItemState = "success"
	toolStateError            ToolItemState = "error"
	toolStateCanceled         ToolItemState = "canceled"
	toolStateAwaitingApproval ToolItemState = "awaiting_approval"
)

// Pre-allocated styles for tool block rendering.
var (
	toolLabelStyle  = lipgloss.NewStyle().Bold(true)
	toolDetailStyle = lipgloss.NewStyle()
	toolOutputStyle = lipgloss.NewStyle().Foreground(tui.Muted).PaddingLeft(4)
)

func formatParamPreview(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, k := range sortedParamKeys(params) {
		parts = append(parts, fmt.Sprintf("%s=%s", sanitizeParamKey(k), formatParamValue(params[k])))
	}
	return strings.Join(parts, "  ")
}

// renderToolBlock renders a tool transcript item with state-specific icon and styling.
func renderToolBlock(toolName string, state ToolItemState, duration, preview, output string, width int) string {
	if width < 10 {
		width = 10
	}
	toolName = sanitizeDisplayText(toolName)

	icon, color := toolStateVisual(state)

	label := toolLabelStyle.Foreground(color).Render(fmt.Sprintf("%s %s", icon, toolName))

	var detail string
	switch state {
	case toolStateRunning:
		detail = toolDetailStyle.Foreground(tui.Muted).Render("running...")
	case toolStateSuccess:
		detail = toolDetailStyle.Foreground(tui.Success).Render(fmt.Sprintf("(%s)", duration))
	case toolStateError:
		detail = toolDetailStyle.Foreground(tui.Error).Render(fmt.Sprintf("failed (%s)", duration))
	case toolStateCanceled:
		detail = toolDetailStyle.Foreground(tui.Muted).Render("canceled")
	case toolStateAwaitingApproval:
		detail = toolDetailStyle.Foreground(tui.Warning).Render("awaiting approval")
	}

	line := ansi.Truncate(fmt.Sprintf(" %s  %s", label, detail), width, "…")

	if preview != "" && (state == toolStateRunning || state == toolStateSuccess || state == toolStateError || state == toolStateCanceled || state == toolStateAwaitingApproval) {
		maxPreview := width - 4
		if maxPreview < 1 {
			maxPreview = 1
		}
		preview = ansi.Truncate(sanitizeDisplayText(preview), maxPreview, "…")
		line += "\n" + ansi.Truncate(toolOutputStyle.Render(preview), width, "…")
	}

	if output != "" && (state == toolStateSuccess || state == toolStateError) {
		maxOutput := width - 4
		if maxOutput < 1 {
			maxOutput = 1
		}
		output = ansi.Truncate(sanitizeDisplayText(output), maxOutput, "…")
		outputLine := ansi.Truncate(toolOutputStyle.Render(output), width, "…")
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
