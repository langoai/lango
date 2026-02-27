package settings

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
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
	si.PromptStyle = lipgloss.NewStyle().Foreground(tui.Primary).Bold(true)
	si.TextStyle = lipgloss.NewStyle().Foreground(tui.Foreground)

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

	// Search bar — always visible
	if m.searching {
		b.WriteString(tui.SearchBarStyle.Render(m.searchInput.View()))
	} else {
		hint := lipgloss.NewStyle().
			Foreground(tui.Dim).
			Italic(true).
			PaddingLeft(1)
		b.WriteString(hint.Render("/ Search..."))
	}
	b.WriteString("\n\n")

	// Menu body
	var body strings.Builder
	if m.searching && m.filtered != nil {
		m.renderFilteredView(&body)
	} else {
		m.renderGroupedView(&body)
	}

	// Wrap in container
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Muted).
		Padding(0, 1)
	b.WriteString(container.Render(body.String()))

	// Help footer with key badges
	b.WriteString("\n")
	if m.searching {
		b.WriteString(tui.HelpBar(
			tui.HelpEntry("↑↓", "Navigate"),
			tui.HelpEntry("Enter", "Select"),
			tui.HelpEntry("Esc", "Cancel"),
		))
	} else {
		b.WriteString(tui.HelpBar(
			tui.HelpEntry("↑↓", "Navigate"),
			tui.HelpEntry("Enter", "Select"),
			tui.HelpEntry("/", "Search"),
			tui.HelpEntry("Esc", "Back"),
		))
	}

	return b.String()
}

func (m MenuModel) renderGroupedView(b *strings.Builder) {
	globalIdx := 0
	for si, section := range m.Sections {
		// Section header
		if section.Title != "" {
			if si > 0 {
				b.WriteString(tui.SeparatorLineStyle.Render("  " + strings.Repeat("─", 38)))
				b.WriteString("\n")
			}
			b.WriteString(tui.SectionHeaderStyle.Render(section.Title))
			b.WriteString("\n")
		} else if si > 0 {
			b.WriteString(tui.SeparatorLineStyle.Render("  " + strings.Repeat("─", 38)))
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
			Foreground(tui.Muted).
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
	const titleWidth = 22
	isSelected := m.Cursor == idx

	cursor := "  "
	titleStyle := lipgloss.NewStyle().Width(titleWidth)
	descStyle := lipgloss.NewStyle().Foreground(tui.Dim)

	if isSelected {
		cursor = tui.CursorStyle.Render("▸ ")
		titleStyle = titleStyle.Foreground(tui.Accent).Bold(true)
		descStyle = descStyle.Foreground(tui.Accent)
	}

	// Handle search highlighting
	title := cat.Title
	desc := cat.Desc

	if m.searching && m.searchInput.Value() != "" {
		query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
		highlightedTitle := m.highlightMatch(title, query, isSelected)
		highlightedDesc := m.highlightMatch(desc, query, isSelected)

		b.WriteString(cursor)
		b.WriteString(lipgloss.NewStyle().Width(titleWidth).Render(highlightedTitle))
		if desc != "" {
			b.WriteString(" ")
			b.WriteString(highlightedDesc)
		}
	} else {
		b.WriteString(cursor)
		b.WriteString(titleStyle.Render(title))
		if desc != "" {
			b.WriteString(descStyle.Render(desc))
		}
	}
	b.WriteString("\n")
}

// highlightMatch highlights matching substrings with amber color.
func (m MenuModel) highlightMatch(text, query string, selected bool) string {
	if query == "" {
		return text
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, query)
	if idx < 0 {
		if selected {
			return lipgloss.NewStyle().Foreground(tui.Accent).Bold(true).Render(text)
		}
		return lipgloss.NewStyle().Foreground(tui.Dim).Render(text)
	}

	matchStyle := lipgloss.NewStyle().Foreground(tui.Warning).Bold(true)
	if selected {
		matchStyle = matchStyle.Underline(true)
	}

	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]

	normalStyle := lipgloss.NewStyle().Foreground(tui.Dim)
	if selected {
		normalStyle = lipgloss.NewStyle().Foreground(tui.Accent).Bold(true)
	}

	return normalStyle.Render(before) + matchStyle.Render(match) + normalStyle.Render(after)
}
