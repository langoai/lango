package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
)

// chatState tracks what the chat model is currently doing.
type chatState int

const (
	stateIdle      chatState = iota
	stateStreaming            // waiting for agent response
	stateApproving            // waiting for approval input
)

// renderStatusBar returns a top status bar showing model, session, and state.
func renderStatusBar(cfg *config.Config, sessionKey string, state chatState, width int) string {
	bg := lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Foreground(tui.Foreground)

	model := cfg.Agent.Provider
	if model == "" {
		model = "default"
	}
	modelBadge := lipgloss.NewStyle().
		Background(tui.Primary).
		Foreground(tui.Foreground).
		Bold(true).
		Padding(0, 1).
		Render(model)

	var stateStr string
	switch state {
	case stateIdle:
		stateStr = lipgloss.NewStyle().Foreground(tui.Success).Render("ready")
	case stateStreaming:
		stateStr = lipgloss.NewStyle().Foreground(tui.Warning).Render("thinking...")
	case stateApproving:
		stateStr = lipgloss.NewStyle().Foreground(tui.Warning).Render("approval needed")
	}

	left := fmt.Sprintf(" %s  %s", modelBadge, stateStr)
	right := lipgloss.NewStyle().Foreground(tui.Muted).Render(fmt.Sprintf("session: %s ", sessionKey))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return bg.Width(width).Render(left + strings.Repeat(" ", gap) + right)
}

// renderHelpBar returns a bottom help bar with context-sensitive key bindings.
func renderHelpBar(state chatState, width int) string {
	var entries []string
	switch state {
	case stateIdle:
		entries = []string{
			tui.HelpEntry("Enter", "send"),
			tui.HelpEntry("Alt+Enter", "newline"),
			tui.HelpEntry("Ctrl+C", "quit"),
			tui.HelpEntry("/help", "commands"),
		}
	case stateStreaming:
		entries = []string{
			tui.HelpEntry("Ctrl+C", "cancel"),
		}
	case stateApproving:
		entries = []string{
			tui.HelpEntry("a", "allow"),
			tui.HelpEntry("s", "allow session"),
			tui.HelpEntry("d/esc", "deny"),
		}
	}
	return tui.HelpBar(entries...)
}
