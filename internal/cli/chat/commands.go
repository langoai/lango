package chat

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// slashCommand defines a chat slash command.
type slashCommand struct {
	name  string
	alias string
	desc  string
	run   func(m *ChatModel, args string) tea.Cmd
}

// slashCommands returns the built-in slash command registry.
func slashCommands() []slashCommand {
	return []slashCommand{
		{
			name: "/help",
			desc: "Show available commands and key bindings",
			run:  cmdHelp,
		},
		{
			name: "/clear",
			desc: "Clear chat and start a new session",
			run:  cmdClear,
		},
		{
			name:  "/new",
			alias: "/clear",
			desc:  "Alias for /clear",
			run:   cmdClear,
		},
		{
			name: "/model",
			desc: "Show current model/provider name",
			run:  cmdModel,
		},
		{
			name: "/status",
			desc: "Show active runtime features",
			run:  cmdStatus,
		},
		{
			name: "/exit",
			desc: "Exit the chat",
			run:  cmdExit,
		},
		{
			name:  "/quit",
			alias: "/exit",
			desc:  "Alias for /exit",
			run:   cmdExit,
		},
	}
}

// dispatchSlash checks if the input is a slash command and executes it.
// Returns true and a tea.Cmd if handled, false otherwise.
func dispatchSlash(m *ChatModel, input string) (bool, tea.Cmd) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return false, nil
	}

	parts := strings.SplitN(input, " ", 2)
	cmd := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	for _, sc := range slashCommands() {
		if sc.name == cmd || (sc.alias != "" && sc.alias == cmd) {
			return true, sc.run(m, args)
		}
	}

	// Unknown command.
	return true, func() tea.Msg {
		return SystemMsg{Text: fmt.Sprintf("Unknown command: %s. Type /help for available commands.", cmd)}
	}
}

func cmdHelp(m *ChatModel, _ string) tea.Cmd {
	var b strings.Builder
	header := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Commands")
	b.WriteString(header + "\n")
	for _, sc := range slashCommands() {
		if sc.alias != "" {
			continue // skip aliases in listing
		}
		name := lipgloss.NewStyle().Bold(true).Foreground(tui.Highlight).Render(sc.name)
		fmt.Fprintf(&b, "  %s  %s\n", name, sc.desc)
	}

	b.WriteString("\n")
	keys := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Key Bindings")
	b.WriteString(keys + "\n")
	bindings := []struct{ key, desc string }{
		{"Enter", "Send message"},
		{"Alt+Enter", "Insert newline"},
		{"Ctrl+C", "Cancel generation / quit"},
		{"Ctrl+D", "Quit immediately"},
		{"PgUp/PgDn", "Scroll chat"},
	}
	for _, kb := range bindings {
		k := lipgloss.NewStyle().Bold(true).Foreground(tui.Highlight).Render(kb.key)
		fmt.Fprintf(&b, "  %s  %s\n", k, kb.desc)
	}

	return func() tea.Msg {
		return SystemMsg{Text: b.String()}
	}
}

func cmdClear(m *ChatModel, _ string) tea.Cmd {
	m.chatView.clear()
	m.sessionKey = generateSessionKey()
	return func() tea.Msg {
		return SystemMsg{Text: "Chat cleared. New session started."}
	}
}

func cmdModel(m *ChatModel, _ string) tea.Cmd {
	provider := m.cfg.Agent.Provider
	if provider == "" {
		provider = "(not configured)"
	}
	model := m.cfg.Agent.Model
	if model == "" {
		model = "(auto)"
	}
	return func() tea.Msg {
		return SystemMsg{Text: fmt.Sprintf("Provider: %s  Model: %s", provider, model)}
	}
}

func cmdStatus(m *ChatModel, _ string) tea.Cmd {
	cfg := m.cfg
	type feature struct {
		name    string
		cfgOn   bool
		runtime string // "active" or "configured but not active in TUI mode"
	}

	activeFeatures := []feature{
		{"Knowledge", cfg.Knowledge.Enabled, "active"},
		{"Embedding & RAG", cfg.Embedding.Provider != "", "active"},
		{"Graph", cfg.Graph.Enabled, "active"},
		{"Obs. Memory", cfg.ObservationalMemory.Enabled, "active"},
	}

	tuiInactive := []feature{
		{"Gateway", cfg.Server.HTTPEnabled, "configured but not active in TUI mode"},
		{"Cron", cfg.Cron.Enabled, "configured but not active in TUI mode"},
		{"MCP", cfg.MCP.Enabled, "configured but not active in TUI mode"},
		{"P2P", cfg.P2P.Enabled, "configured but not active in TUI mode"},
		{"Payment", cfg.Payment.Enabled, "configured but not active in TUI mode"},
	}

	var b strings.Builder
	header := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).Render("Runtime Status (TUI Mode)")
	b.WriteString(header + "\n")

	for _, f := range activeFeatures {
		if f.cfgOn {
			fmt.Fprintf(&b, "  %s %s\n", tui.FormatPass(f.name), lipgloss.NewStyle().Foreground(tui.Muted).Render(f.runtime))
		}
	}
	for _, f := range tuiInactive {
		if f.cfgOn {
			fmt.Fprintf(&b, "  %s %s\n", tui.FormatWarn(f.name), lipgloss.NewStyle().Foreground(tui.Muted).Render(f.runtime))
		}
	}

	return func() tea.Msg {
		return SystemMsg{Text: b.String()}
	}
}

func cmdExit(_ *ChatModel, _ string) tea.Cmd {
	return tea.Quit
}
