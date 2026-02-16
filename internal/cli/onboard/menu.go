package onboard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Category struct {
	ID    string
	Title string
	Desc  string
}

type MenuModel struct {
	Categories []Category
	Cursor     int
	Selected   string
	Width      int
	Height     int
}

func NewMenuModel() MenuModel {
	return MenuModel{
		Categories: []Category{
			{"providers", "☁️ Providers", "Manage multi-provider configurations"},
			{"agent", "🤖 Agent", "Provider, Model, Key"},
			{"server", "🌐 Server", "Host, Port, Networking"},
			{"channels", "📡 Channels", "Telegram, Discord, Slack"},
			{"tools", "🛠️ Tools", "Exec, Browser, Filesystem"},
			{"session", "📂 Session", "Database, TTL, History"},
			{"security", "🔒 Security", "PII, Approval, Encryption"},
			{"auth", "🔑 Auth", "OIDC provider configuration"},
			{"knowledge", "🧠 Knowledge", "Learning, Skills, Context limits"},
			{"observational_memory", "🔬 Observational Memory", "Observer, Reflector, Thresholds"},
			{"embedding", "🔗 Embedding & RAG", "Provider, Model, RAG settings"},
			{"graph", "📊 Graph Store", "Knowledge graph, GraphRAG settings"},
			{"multi_agent", "🔀 Multi-Agent", "Orchestration mode"},
			{"a2a", "🌍 A2A Protocol", "Agent-to-Agent, remote agents"},
			{"save", "💾 Save & Exit", "Save encrypted profile"},
			{"cancel", "❌ Cancel", "Exit without saving"},
		},
		Cursor: 0,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

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
			cursor = "▸ "
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
