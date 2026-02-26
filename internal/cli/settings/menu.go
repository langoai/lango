package settings

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Category represents a configuration category in the menu.
type Category struct {
	ID    string
	Title string
	Desc  string
}

// MenuModel manages the configuration menu.
type MenuModel struct {
	Categories []Category
	Cursor     int
	Selected   string
	Width      int
	Height     int
}

// NewMenuModel creates a new menu model with all configuration categories.
func NewMenuModel() MenuModel {
	return MenuModel{
		Categories: []Category{
			{"providers", "Providers", "Manage multi-provider configurations"},
			{"agent", "Agent", "Provider, Model, Key"},
			{"server", "Server", "Host, Port, Networking"},
			{"channels", "Channels", "Telegram, Discord, Slack"},
			{"tools", "Tools", "Exec, Browser, Filesystem"},
			{"session", "Session", "Database, TTL, History"},
			{"security", "Security", "PII, Approval, Encryption"},
			{"auth", "Auth", "OIDC provider configuration"},
			{"knowledge", "Knowledge", "Learning, Context limits"},
			{"skill", "Skill", "File-based skill system"},
			{"observational_memory", "Observational Memory", "Observer, Reflector, Thresholds"},
			{"embedding", "Embedding & RAG", "Provider, Model, RAG settings"},
			{"graph", "Graph Store", "Knowledge graph, GraphRAG settings"},
			{"multi_agent", "Multi-Agent", "Orchestration mode"},
			{"a2a", "A2A Protocol", "Agent-to-Agent, remote agents"},
			{"payment", "Payment", "Blockchain wallet, spending limits, X402"},
			{"cron", "Cron Scheduler", "Scheduled jobs, timezone, history"},
			{"background", "Background Tasks", "Async tasks, concurrency limits"},
			{"workflow", "Workflow Engine", "DAG workflows, timeouts, state"},
			{"librarian", "Librarian", "Proactive knowledge extraction, inquiries"},
			{"p2p", "P2P Network", "Peer-to-peer networking, discovery, handshake"},
			{"p2p_zkp", "P2P ZKP", "Zero-knowledge proof settings"},
			{"p2p_pricing", "P2P Pricing", "Paid tool invocations"},
			{"p2p_owner", "P2P Owner Protection", "Owner PII leak prevention"},
			{"p2p_sandbox", "P2P Sandbox", "Tool isolation, container sandbox"},
			{"security_keyring", "Security Keyring", "OS keyring for passphrase storage"},
			{"security_db", "Security DB Encryption", "SQLCipher database encryption"},
			{"security_kms", "Security KMS", "Cloud KMS / HSM backends"},
			{"save", "Save & Exit", "Save encrypted profile"},
			{"cancel", "Cancel", "Exit without saving"},
		},
		Cursor: 0,
	}
}

// Init implements tea.Model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for the menu.
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Categories)-1 {
				m.Cursor++
			}
		case "enter":
			m.Selected = m.Categories[m.Cursor].ID
			return m, nil
		}
	}
	return m, nil
}

// View renders the configuration menu.
func (m MenuModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Configuration Menu"))
	b.WriteString("\n\n")

	for i, cat := range m.Categories {
		cursor := "  "
		itemStyle := lipgloss.NewStyle()

		if m.Cursor == i {
			cursor = "\u25b8 "
			itemStyle = itemStyle.Foreground(lipgloss.Color("#04B575")).Bold(true)
		}

		b.WriteString(cursor)
		b.WriteString(itemStyle.Render(cat.Title))
		if cat.Desc != "" {
			b.WriteString(" " + lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(cat.Desc))
		}
		b.WriteString("\n")
	}

	return b.String()
}
