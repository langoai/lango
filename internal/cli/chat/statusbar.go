package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
)

func renderShellBar(left, right string, width int, bg lipgloss.Color) string {
	w := max(width, 1)
	if lipgloss.Width(left) >= w {
		left = ansi.Truncate(left, w, "…")
		right = ""
	} else if right != "" {
		maxRight := w - lipgloss.Width(left) - 1
		if maxRight < 0 {
			maxRight = 0
		}
		right = ansi.Truncate(right, maxRight, "…")
	}

	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	content := left + strings.Repeat(" ", gap) + right
	content = ansi.Truncate(content, w, "…")

	return lipgloss.NewStyle().
		Background(bg).
		Foreground(tui.Foreground).
		Width(w).
		Render(content)
}

// chatState tracks the current operator-visible TUI turn state.
type chatState int

const (
	stateIdle chatState = iota
	stateStreaming
	stateApproving
	stateCancelling
	stateFailed
)

func renderHeader(cfg *config.Config, sessionKey string, width int) string {
	return renderHeaderWithSetup(cfg, sessionKey, width, false)
}

func renderHeaderWithSetup(cfg *config.Config, sessionKey string, width int, setupRequired bool) string {
	productBadge := lipgloss.NewStyle().
		Background(tui.Primary).
		Foreground(tui.Foreground).
		Bold(true).
		Padding(0, 1).
		Render("Lango")

	provider := ""
	model := ""
	if cfg != nil {
		provider = singleLineValue(ansi.Strip(cfg.Agent.Provider))
		model = singleLineValue(ansi.Strip(cfg.Agent.Model))
	}
	if provider == "" {
		provider = "default"
	}
	if model == "" {
		model = "auto"
	}
	sessionKey = singleLineValue(ansi.Strip(sessionKey))

	display := fmt.Sprintf("%s · %s", provider, model)
	color := tui.Foreground
	if setupRequired {
		display = setupRequiredLabel
		color = tui.Warning
	}
	modelText := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(display)

	left := fmt.Sprintf(" %s  %s", productBadge, modelText)
	right := lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render(fmt.Sprintf("session: %s ", sessionKey))

	return renderShellBar(left, right, width, lipgloss.Color("#132238"))
}

func renderTurnStrip(state chatState, width int) string {
	return renderTurnStripWithSetup(state, width, false)
}

func renderTurnStripWithSetup(state chatState, width int, setupRequired bool) string {
	label, hint, color := turnStateCopyWithSetup(state, setupRequired)
	left := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(" " + label)
	right := lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render(hint + " ")

	return renderShellBar(left, right, width, lipgloss.Color("#0f1724"))
}

func renderHelpBar(state chatState, width int) string {
	return renderHelpBarWithSetup(state, width, false)
}

func renderHelpBarWithSetup(state chatState, width int, setupRequired bool) string {
	w := max(width, 1)
	var entries []string
	switch state {
	case stateIdle, stateFailed:
		if setupRequired {
			entries = []string{
				tui.HelpEntry("lango onboard", "setup"),
				tui.HelpEntry("lango settings", "config"),
				tui.HelpEntry("lango doctor", "check"),
				tui.HelpEntry("/help", "commands"),
				tui.HelpEntry("Ctrl+D", "quit"),
			}
		} else {
			entries = []string{
				tui.HelpEntry("Enter", "send"),
				tui.HelpEntry("Alt+Enter", "newline"),
				tui.HelpEntry("Ctrl+C", "quit x2"),
				tui.HelpEntry("Ctrl+D", "quit"),
				tui.HelpEntry("/help", "commands"),
			}
		}
	case stateStreaming:
		entries = []string{
			tui.HelpEntry("Ctrl+C", "cancel"),
			tui.HelpEntry("Ctrl+D", "quit"),
		}
	case stateApproving:
		entries = []string{
			tui.HelpEntry("a", "allow"),
			tui.HelpEntry("s", "allow session"),
			tui.HelpEntry("d/Esc", "deny"),
			tui.HelpEntry("Ctrl+D", "quit"),
		}
	case stateCancelling:
		entries = []string{
			tui.HelpEntry("Ctrl+D", "quit"),
		}
	}
	bar := tui.HelpBar(entries...)
	return ansi.Truncate(bar, w, "")
}

func turnStateCopy(state chatState) (label, hint string, color lipgloss.Color) {
	return turnStateCopyWithSetup(state, false)
}

func turnStateCopyWithSetup(state chatState, setupRequired bool) (label, hint string, color lipgloss.Color) {
	if setupRequired && (state == stateIdle || state == stateFailed) {
		return setupRequiredLabel, setupRequiredGuidance, tui.Warning
	}
	switch state {
	case stateIdle:
		return "Ready", "Enter sends · /help shows commands", tui.Success
	case stateStreaming:
		return "Streaming", "Ctrl+C cancels the current turn", tui.Warning
	case stateApproving:
		return "Approval Required", "Choose a / s / d/Esc · Ctrl+D quits", tui.Warning
	case stateCancelling:
		return "Cancelling", "Waiting for the current turn to stop", tui.Muted
	case stateFailed:
		return "Last Turn Failed", "Type to retry or inspect /status", tui.Error
	default:
		return "Ready", "", tui.Success
	}
}
