package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
)

// renderApprovalDialog renders a Tier 2 fullscreen approval dialog overlay.
func renderApprovalDialog(vm approval.ApprovalViewModel, state *approvalState, width, height int) string {
	if state == nil {
		state = &approvalState{}
	}
	isConfirmPending := state != nil && state.confirmPending
	confirmKey := "a"
	if state != nil && strings.TrimSpace(state.confirmAction) != "" {
		confirmKey = state.confirmAction
	}
	scrollOffset := state.scrollOffset
	splitMode := state.splitMode

	dialogWidth := width - 4
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	contentHeight := height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Header: risk badge + tool name.
	safeRiskLevel := strings.ToUpper(sanitizeDisplayText(vm.Risk.Level))
	riskColor := riskLevelColor(strings.ToLower(safeRiskLevel))
	safeToolName := sanitizeDisplayText(vm.Request.ToolName)
	safeRiskLabel := sanitizeDisplayText(vm.Risk.Label)
	riskBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Foreground).
		Background(riskColor).
		Padding(0, 1).
		Render(safeRiskLevel)

	toolName := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Highlight).
		Render(safeToolName)

	header := fmt.Sprintf(" %s  %s  %s", riskBadge, toolName, safeRiskLabel)

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
		summary = fmt.Sprintf("Execute tool: %s", safeToolName)
	}
	summary = sanitizeDisplayText(summary)
	summaryBlock := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(tui.Foreground).
		Render(summary)

	// Rule explanation.
	var explanationBlock string
	if vm.RuleExplanation != "" {
		explanation := sanitizeDisplayText(vm.RuleExplanation)
		explanationBlock = lipgloss.NewStyle().
			Foreground(tui.Muted).
			Italic(true).
			PaddingLeft(2).
			Render("Why: " + explanation)
	}

	// Parameters.
	var paramsBlock string
	if len(vm.Request.Params) > 0 {
		var parts []string
		for _, k := range sortedParamKeys(vm.Request.Params) {
			val := formatParamValue(vm.Request.Params[k])
			if len(val) > 120 {
				val = val[:117] + "..."
			}
			parts = append(parts, fmt.Sprintf("  %s: %s", sanitizeParamKey(k), val))
		}
		paramsBlock = lipgloss.NewStyle().
			Foreground(tui.Muted).
			Render(strings.Join(parts, "\n"))
	}

	// Diff preview (if available).
	var diffBlock string
	hasScrollableDiff := false
	if vm.DiffContent != "" {
		allLines := strings.Split(vm.DiffContent, "\n")

		// Use cached styled lines if cache key matches; otherwise build and cache.
		cachedLines := state.diffCache.lines
		if cachedLines == nil ||
			state.diffCache.content != vm.DiffContent ||
			state.diffCache.width != width ||
			state.diffCache.splitMode != splitMode {
			cachedLines = make([]string, 0, len(allLines))
			for _, line := range allLines {
				line = ansi.Strip(line)
				switch {
				case strings.HasPrefix(line, "+"):
					cachedLines = append(cachedLines, lipgloss.NewStyle().Foreground(tui.Success).Render(line))
				case strings.HasPrefix(line, "-"):
					cachedLines = append(cachedLines, lipgloss.NewStyle().Foreground(tui.Error).Render(line))
				case strings.HasPrefix(line, "@@"):
					cachedLines = append(cachedLines, lipgloss.NewStyle().Foreground(tui.Info).Render(line))
				default:
					cachedLines = append(cachedLines, line)
				}
			}
			state.diffCache = diffLineCache{
				content:   vm.DiffContent,
				width:     width,
				splitMode: splitMode,
				lines:     cachedLines,
			}
		}

		// Apply scroll offset over the cached styled lines.
		start := scrollOffset
		if start >= len(cachedLines) {
			start = max(len(cachedLines)-1, 0)
		}
		if start < 0 {
			start = 0
		}

		// Limit visible lines.
		visible := contentHeight - 8
		if visible < 3 {
			visible = 3
		}
		maxStart := len(cachedLines) - visible
		if maxStart < 0 {
			maxStart = 0
		}
		if state.scrollOffset > maxStart {
			state.scrollOffset = maxStart
			start = maxStart
		}
		hasScrollableDiff = len(cachedLines) > visible
		end := start + visible
		if end > len(cachedLines) {
			end = len(cachedLines)
		}

		visibleLines := cachedLines[start:end]

		diffMode := "unified"
		if splitMode {
			diffMode = "split"
		}
		diffHeader := lipgloss.NewStyle().Bold(true).Foreground(tui.Muted).
			Render(fmt.Sprintf("  Diff [%s] (%d/%d lines)", diffMode, end, len(allLines)))
		diffBody := lipgloss.NewStyle().PaddingLeft(2).
			Render(strings.Join(visibleLines, "\n"))
		diffBlock = diffHeader + "\n" + diffBody
	}

	// Action bar.
	var actionBar string
	if isConfirmPending {
		actionBar = lipgloss.NewStyle().
			Bold(true).
			Foreground(tui.Warning).
			Render(fmt.Sprintf("  Press '%s' again to confirm (destructive operation)  d/Esc denies", confirmKey))
	} else if vm.DiffContent != "" {
		entries := []string{
			tui.HelpEntry("a", "allow"),
			tui.HelpEntry("s", "allow session"),
			tui.HelpEntry("d/Esc", "deny"),
			tui.HelpEntry("t", "split"),
		}
		if hasScrollableDiff {
			entries = append(entries, tui.HelpEntry("\u2191\u2193", "scroll"))
		}
		actionBar = tui.HelpBar(entries...)
	} else {
		actionBar = tui.HelpBar(
			tui.HelpEntry("a", "allow"),
			tui.HelpEntry("s", "allow session"),
			tui.HelpEntry("d/Esc", "deny"),
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
func handleApprovalDialogKey(msg tea.KeyMsg, state *approvalState) tea.Cmd {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		state.ScrollDiff(-3)
		return nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		state.ScrollDiff(3)
		return nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
		state.ToggleSplit()
		return nil
	}
	return nil
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
