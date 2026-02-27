package settings

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Category represents a configuration category in the menu.
type Category struct {
	ID    string
	Title string
	Desc  string
}

// Section groups related categories under a heading.
type Section struct {
	Title      string
	Categories []Category
}

// MenuModel manages the configuration menu.
type MenuModel struct {
	Sections []Section
	Cursor   int
	Selected string
	Width    int
	Height   int

	// Search
	searching   bool
	searchInput textinput.Model
	filtered    []Category // filtered results (nil when not searching)
}

// allCategories returns a flat list of all selectable categories across sections.
func (m *MenuModel) allCategories() []Category {
	var all []Category
	for _, s := range m.Sections {
		all = append(all, s.Categories...)
	}
	return all
}

// AllCategories returns a flat list of all categories (public, for tests).
func (m MenuModel) AllCategories() []Category {
	return m.allCategories()
}

// IsSearching returns true when the menu is in search mode.
func (m MenuModel) IsSearching() bool {
	return m.searching
}

// selectableItems returns the list the cursor currently navigates.
func (m *MenuModel) selectableItems() []Category {
	if m.searching && m.filtered != nil {
		return m.filtered
	}
	return m.allCategories()
}

// NewMenuModel creates a new menu model with grouped configuration categories.
func NewMenuModel() MenuModel {
	si := textinput.New()
	si.Placeholder = "Type to search..."
	si.CharLimit = 40
	si.Width = 30
	si.Prompt = "/ "
	si.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	si.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))

	return MenuModel{
		Sections: []Section{
			{
				Title: "Core",
				Categories: []Category{
					{"providers", "Providers", "Multi-provider configurations"},
					{"agent", "Agent", "Provider, Model, Key"},
					{"server", "Server", "Host, Port, Networking"},
					{"session", "Session", "Database, TTL, History"},
				},
			},
			{
				Title: "Communication",
				Categories: []Category{
					{"channels", "Channels", "Telegram, Discord, Slack"},
					{"tools", "Tools", "Exec, Browser, Filesystem"},
					{"multi_agent", "Multi-Agent", "Orchestration mode"},
					{"a2a", "A2A Protocol", "Agent-to-Agent, remote agents"},
				},
			},
			{
				Title: "AI & Knowledge",
				Categories: []Category{
					{"knowledge", "Knowledge", "Learning, Context limits"},
					{"skill", "Skill", "File-based skill system"},
					{"observational_memory", "Observational Memory", "Observer, Reflector, Thresholds"},
					{"embedding", "Embedding & RAG", "Provider, Model, RAG settings"},
					{"graph", "Graph Store", "Knowledge graph, GraphRAG settings"},
					{"librarian", "Librarian", "Proactive knowledge extraction"},
				},
			},
			{
				Title: "Infrastructure",
				Categories: []Category{
					{"payment", "Payment", "Blockchain wallet, spending limits"},
					{"cron", "Cron Scheduler", "Scheduled jobs, timezone, history"},
					{"background", "Background Tasks", "Async tasks, concurrency limits"},
					{"workflow", "Workflow Engine", "DAG workflows, timeouts, state"},
				},
			},
			{
				Title: "P2P Network",
				Categories: []Category{
					{"p2p", "P2P Network", "Peer-to-peer networking, discovery"},
					{"p2p_zkp", "P2P ZKP", "Zero-knowledge proof settings"},
					{"p2p_pricing", "P2P Pricing", "Paid tool invocations"},
					{"p2p_owner", "P2P Owner Protection", "Owner PII leak prevention"},
					{"p2p_sandbox", "P2P Sandbox", "Tool isolation, container sandbox"},
				},
			},
			{
				Title: "Security",
				Categories: []Category{
					{"security", "Security", "PII, Approval, Encryption"},
					{"auth", "Auth", "OIDC provider configuration"},
					{"security_keyring", "Security Keyring", "OS keyring for passphrase storage"},
					{"security_db", "Security DB Encryption", "SQLCipher database encryption"},
					{"security_kms", "Security KMS", "Cloud KMS / HSM backends"},
				},
			},
			{
				Title: "",
				Categories: []Category{
					{"save", "Save & Exit", "Save encrypted profile"},
					{"cancel", "Cancel", "Exit without saving"},
				},
			},
		},
		Cursor:      0,
		searchInput: si,
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
		key := msg.String()

		// --- Search mode handling ---
		if m.searching {
			switch key {
			case "esc":
				m.searching = false
				m.filtered = nil
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				m.Cursor = 0
				return m, nil
			case "enter":
				items := m.selectableItems()
				if len(items) > 0 && m.Cursor < len(items) {
					m.Selected = items[m.Cursor].ID
					m.searching = false
					m.filtered = nil
					m.searchInput.SetValue("")
					m.searchInput.Blur()
				}
				return m, nil
			case "up", "shift+tab":
				if m.Cursor > 0 {
					m.Cursor--
				}
				return m, nil
			case "down", "tab":
				items := m.selectableItems()
				if m.Cursor < len(items)-1 {
					m.Cursor++
				}
				return m, nil
			default:
				// Forward to text input
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.applyFilter()
				return m, cmd
			}
		}

		// --- Normal mode handling ---
		switch key {
		case "/":
			m.searching = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			m.Cursor = 0
			return m, textinput.Blink
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			items := m.selectableItems()
			if m.Cursor < len(items)-1 {
				m.Cursor++
			}
		case "enter":
			items := m.selectableItems()
			if len(items) > 0 && m.Cursor < len(items) {
				m.Selected = items[m.Cursor].ID
			}
			return m, nil
		}
	}
	return m, nil
}

// applyFilter updates the filtered list based on the current search query.
func (m *MenuModel) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		m.filtered = nil
		m.Cursor = 0
		return
	}

	var results []Category
	all := m.allCategories()
	for _, cat := range all {
		title := strings.ToLower(cat.Title)
		desc := strings.ToLower(cat.Desc)
		id := strings.ToLower(cat.ID)
		if strings.Contains(title, query) || strings.Contains(desc, query) || strings.Contains(id, query) {
			results = append(results, cat)
		}
	}
	m.filtered = results
	m.Cursor = 0
}

// View renders the configuration menu.
func (m MenuModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Configuration Menu"))
	b.WriteString("\n\n")

	// Search bar
	if m.searching {
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1).
			MarginBottom(1)
		b.WriteString(boxStyle.Render(m.searchInput.View()))
		b.WriteString("\n\n")
	} else {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
		b.WriteString(hintStyle.Render("Press / to search"))
		b.WriteString("\n\n")
	}

	if m.searching && m.filtered != nil {
		// Render filtered results
		m.renderFilteredView(&b)
	} else {
		// Render grouped view
		m.renderGroupedView(&b)
	}

	// Help footer
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	if m.searching {
		b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: cancel search"))
	} else {
		b.WriteString(helpStyle.Render("↑/↓/j/k: navigate • enter: select • /: search • esc: quit"))
	}

	return b.String()
}

func (m MenuModel) renderGroupedView(b *strings.Builder) {
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Bold(true).
		MarginTop(1)
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151"))

	globalIdx := 0
	for si, section := range m.Sections {
		// Section header
		if section.Title != "" {
			if si > 0 {
				b.WriteString(separatorStyle.Render("  ────────────────────────────────────"))
				b.WriteString("\n")
			}
			b.WriteString(sectionStyle.Render("  " + section.Title))
			b.WriteString("\n")
		} else if si > 0 {
			b.WriteString(separatorStyle.Render("  ────────────────────────────────────"))
			b.WriteString("\n")
		}

		for _, cat := range section.Categories {
			m.renderItem(b, cat, globalIdx)
			globalIdx++
		}
	}
}

func (m MenuModel) renderFilteredView(b *strings.Builder) {
	if len(m.filtered) == 0 {
		noResult := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)
		b.WriteString(noResult.Render("  No matching items"))
		b.WriteString("\n")
		return
	}

	for i, cat := range m.filtered {
		m.renderItem(b, cat, i)
	}
}

func (m MenuModel) renderItem(b *strings.Builder, cat Category, idx int) {
	cursor := "  "
	titleStyle := lipgloss.NewStyle()
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	if m.Cursor == idx {
		cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render("▸ ")
		titleStyle = titleStyle.Foreground(lipgloss.Color("#04B575")).Bold(true)
		descStyle = descStyle.Foreground(lipgloss.Color("#04B575"))
	}

	// Highlight search matches in title/desc
	title := cat.Title
	desc := cat.Desc

	if m.searching && m.searchInput.Value() != "" {
		query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
		title = m.highlightMatch(title, query, m.Cursor == idx)
		desc = m.highlightMatch(desc, query, m.Cursor == idx)
	}

	b.WriteString(cursor)
	if m.searching && m.searchInput.Value() != "" {
		// Already highlighted, render raw
		b.WriteString(title)
	} else {
		b.WriteString(titleStyle.Render(title))
	}
	if desc != "" {
		b.WriteString(" ")
		if m.searching && m.searchInput.Value() != "" {
			b.WriteString(desc)
		} else {
			b.WriteString(descStyle.Render(desc))
		}
	}
	b.WriteString("\n")
}

// highlightMatch highlights matching substrings with a yellow/amber color.
func (m MenuModel) highlightMatch(text, query string, selected bool) string {
	if query == "" {
		return text
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, query)
	if idx < 0 {
		if selected {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true).Render(text)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(text)
	}

	matchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)
	if selected {
		matchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true).Underline(true)
	}

	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]

	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	if selected {
		normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	}

	return normalStyle.Render(before) + matchStyle.Render(match) + normalStyle.Render(after)
}
