package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
)

// dialogScrollOffset tracks the diff viewport scroll position (package-level for simplicity).
var dialogScrollOffset int

// dialogSplitMode toggles between unified (false) and split (true) diff display.
var dialogSplitMode bool

// renderApprovalDialog renders a Tier 2 fullscreen approval dialog overlay.
func renderApprovalDialog(vm approval.ApprovalViewModel, width, height int, confirmPending ...bool) string {
	isConfirmPending := len(confirmPending) > 0 && confirmPending[0]
	dialogWidth := width - 4
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	contentHeight := height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Header: risk badge + tool name.
	riskColor := riskLevelColor(vm.Risk.Level)
	riskBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Foreground).
		Background(riskColor).
		Padding(0, 1).
		Render(strings.ToUpper(vm.Risk.Level))

	toolName := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Highlight).
		Render(vm.Request.ToolName)

	header := fmt.Sprintf(" %s  %s  %s", riskBadge, toolName, vm.Risk.Label)

	// Channel origin (if from a channel session).
	var originLine string
	if origin := formatChannelOrigin(vm.Request.SessionKey); origin != "" {
		originLine = lipgloss.NewStyle().
			Foreground(tui.Info).
			PaddingLeft(2).
			Render("← " + origin)
	}

	// Summary.
	summary := vm.Request.Summary
	if summary == "" {
		summary = fmt.Sprintf("Execute tool: %s", vm.Request.ToolName)
	}
	summaryBlock := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(tui.Foreground).
		Render(summary)

	// Rule explanation.
	var explanationBlock string
	if vm.RuleExplanation != "" {
		explanationBlock = lipgloss.NewStyle().
			Foreground(tui.Muted).
			Italic(true).
			PaddingLeft(2).
			Render("Why: " + vm.RuleExplanation)
	}

	// Parameters.
	var paramsBlock string
	if len(vm.Request.Params) > 0 {
		var parts []string
		for k, v := range vm.Request.Params {
			val := fmt.Sprintf("%v", v)
			if len(val) > 120 {
				val = val[:117] + "..."
			}
			parts = append(parts, fmt.Sprintf("  %s: %s", k, val))
		}
		paramsBlock = lipgloss.NewStyle().
			Foreground(tui.Muted).
			Render(strings.Join(parts, "\n"))
	}

	// Diff preview (if available).
	var diffBlock string
	if vm.DiffContent != "" {
		lines := strings.Split(vm.DiffContent, "\n")

		// Apply scroll offset.
		start := dialogScrollOffset
		if start >= len(lines) {
			start = max(len(lines)-1, 0)
		}
		if start < 0 {
			start = 0
		}

		// Limit visible lines.
		visible := contentHeight - 8
		if visible < 3 {
			visible = 3
		}
		end := start + visible
		if end > len(lines) {
			end = len(lines)
		}

		visibleLines := lines[start:end]
		var styledLines []string
		for _, line := range visibleLines {
			switch {
			case strings.HasPrefix(line, "+"):
				styledLines = append(styledLines, lipgloss.NewStyle().Foreground(tui.Success).Render(line))
			case strings.HasPrefix(line, "-"):
				styledLines = append(styledLines, lipgloss.NewStyle().Foreground(tui.Error).Render(line))
			case strings.HasPrefix(line, "@@"):
				styledLines = append(styledLines, lipgloss.NewStyle().Foreground(tui.Info).Render(line))
			default:
				styledLines = append(styledLines, line)
			}
		}

		diffMode := "unified"
		if dialogSplitMode {
			diffMode = "split"
		}
		diffHeader := lipgloss.NewStyle().Bold(true).Foreground(tui.Muted).
			Render(fmt.Sprintf("  Diff [%s] (%d/%d lines)", diffMode, end, len(lines)))
		diffBody := lipgloss.NewStyle().PaddingLeft(2).
			Render(strings.Join(styledLines, "\n"))
		diffBlock = diffHeader + "\n" + diffBody
	}

	// Action bar.
	var actionBar string
	if isConfirmPending {
		actionBar = lipgloss.NewStyle().
			Bold(true).
			Foreground(tui.Warning).
			Render("  Press 'a' again to confirm (destructive operation)")
	} else if vm.DiffContent != "" {
		actionBar = tui.HelpBar(
			tui.HelpEntry("a", "allow"),
			tui.HelpEntry("s", "session"),
			tui.HelpEntry("d/esc", "deny"),
			tui.HelpEntry("\u2191\u2193", "scroll"),
		)
	} else {
		actionBar = tui.HelpBar(
			tui.HelpEntry("a", "allow"),
			tui.HelpEntry("s", "allow session"),
			tui.HelpEntry("d/esc", "deny"),
		)
	}

	// Assemble content.
	sections := []string{header}
	if originLine != "" {
		sections = append(sections, originLine)
	}
	sections = append(sections, "", summaryBlock)
	if explanationBlock != "" {
		sections = append(sections, explanationBlock)
	}
	if paramsBlock != "" {
		sections = append(sections, paramsBlock)
	}
	if diffBlock != "" {
		sections = append(sections, "", diffBlock)
	}
	sections = append(sections, "", actionBar)

	content := strings.Join(sections, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Warning).
		Width(dialogWidth).
		Padding(1, 2).
		Render(content)
}

// handleApprovalDialogKey handles key events for the Tier 2 approval dialog.
func handleApprovalDialogKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		scrollApprovalDialog(-3)
		return nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		scrollApprovalDialog(3)
		return nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
		dialogSplitMode = !dialogSplitMode
		return nil
	}
	return nil
}

// scrollApprovalDialog adjusts the diff viewport scroll position.
func scrollApprovalDialog(delta int) {
	dialogScrollOffset += delta
	if dialogScrollOffset < 0 {
		dialogScrollOffset = 0
	}
}

func riskLevelColor(level string) lipgloss.Color {
	switch level {
	case "critical":
		return tui.Error
	case "high":
		return tui.Warning
	case "moderate":
		return tui.Highlight
	default:
		return tui.Muted
	}
}
