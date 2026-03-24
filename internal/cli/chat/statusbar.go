package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
)

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
	headerWidth := max(width, 1)

	productBadge := lipgloss.NewStyle().
		Background(tui.Primary).
		Foreground(tui.Foreground).
		Bold(true).
		Padding(0, 1).
		Render("Lango")

	provider := strings.TrimSpace(cfg.Agent.Provider)
	if provider == "" {
		provider = "default"
	}
	model := strings.TrimSpace(cfg.Agent.Model)
	if model == "" {
		model = "auto"
	}

	modelText := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Foreground).
		Render(fmt.Sprintf("%s · %s", provider, model))

	left := fmt.Sprintf(" %s  %s", productBadge, modelText)
	right := lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render(fmt.Sprintf("session: %s ", sessionKey))

	gap := headerWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#132238")).
		Foreground(tui.Foreground).
		Width(headerWidth).
		Render(left + strings.Repeat(" ", gap) + right)
}

func renderTurnStrip(state chatState, width int) string {
	stripWidth := max(width, 1)

	label, hint, color := turnStateCopy(state)
	left := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(" " + label)
	right := lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render(hint + " ")

	gap := stripWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0f1724")).
		Foreground(tui.Foreground).
		Width(stripWidth).
		Render(left + strings.Repeat(" ", gap) + right)
}

func renderHelpBar(state chatState, _ int) string {
	var entries []string
	switch state {
	case stateIdle, stateFailed:
		entries = []string{
			tui.HelpEntry("Enter", "send"),
			tui.HelpEntry("Alt+Enter", "newline"),
			tui.HelpEntry("Ctrl+C", "quit"),
			tui.HelpEntry("/help", "commands"),
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
			tui.HelpEntry("d/esc", "deny"),
		}
	case stateCancelling:
		entries = []string{
			tui.HelpEntry("Ctrl+D", "quit"),
		}
	}
	return tui.HelpBar(entries...)
}

func turnStateCopy(state chatState) (label, hint string, color lipgloss.Color) {
	switch state {
	case stateIdle:
		return "Ready", "Enter sends · /help shows commands", tui.Success
	case stateStreaming:
		return "Streaming", "Ctrl+C cancels the current turn", tui.Warning
	case stateApproving:
		return "Approval Required", "Review the tool action and choose a / s / d", tui.Warning
	case stateCancelling:
		return "Cancelling", "Waiting for the current turn to stop", tui.Muted
	case stateFailed:
		return "Last Turn Failed", "Type to retry or inspect /status", tui.Error
	default:
		return "Ready", "", tui.Success
	}
}
