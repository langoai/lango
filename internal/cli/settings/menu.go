package settings

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// Tier constants for category visibility.
const (
	TierBasic    = 0
	TierAdvanced = 1
)

// Category represents a configuration category in the menu.
type Category struct {
	ID    string
	Title string
	Desc  string
	Tier  int // TierBasic or TierAdvanced
}

// Section groups related categories under a heading.
type Section struct {
	Title      string
	Categories []Category
}

// MenuModel manages the configuration menu.
type MenuModel struct {
	Sections     []Section
	Cursor       int
	Selected     string
	Width        int
	Height       int
	showAdvanced bool

	// Search
	searching   bool
	searchInput textinput.Model
	filtered    []Category // filtered results (nil when not searching)

	// Checkers for smart filters
	DirtyChecker   func(string) bool // returns true if category config has been modified
	EnabledChecker func(string) bool // returns true if category feature is enabled
}

// allCategories returns a flat list of all selectable categories across sections.
func (m *MenuModel) allCategories() []Category {
	var all []Category
	for _, s := range m.Sections {
		all = append(all, s.Categories...)
	}
	return all
}

// visibleCategories returns categories respecting the current tier filter.
func (m *MenuModel) visibleCategories() []Category {
	var vis []Category
	for _, s := range m.Sections {
		for _, c := range s.Categories {
			if m.showAdvanced || c.Tier == TierBasic {
				vis = append(vis, c)
			}
		}
	}
	return vis
}

// AllCategories returns a flat list of all categories (public, for tests).
func (m MenuModel) AllCategories() []Category {
	return m.allCategories()
}

// IsSearching returns true when the menu is in search mode.
func (m MenuModel) IsSearching() bool {
	return m.searching
}

// ShowAdvanced returns the current advanced mode state.
func (m MenuModel) ShowAdvanced() bool {
	return m.showAdvanced
}

// selectableItems returns the list the cursor currently navigates.
func (m *MenuModel) selectableItems() []Category {
	if m.searching && m.filtered != nil {
		return m.filtered
	}
	return m.visibleCategories()
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
		showAdvanced: true,
		Sections: []Section{
			{
				Title: "Core",
				Categories: []Category{
					{"providers", "Providers", "Multi-provider configurations", TierBasic},
					{"agent", "Agent", "Provider, Model, Key", TierBasic},
					{"channels", "Channels", "Telegram, Discord, Slack", TierBasic},
					{"tools", "Tools", "Exec, Browser, Filesystem", TierBasic},
					{"server", "Server", "Host, Port, Networking", TierAdvanced},
					{"session", "Session", "Database, TTL, History", TierAdvanced},
					{"logging", "Logging", "Level, Format, Output path", TierAdvanced},
					{"gatekeeper", "Gatekeeper", "Response sanitization filters", TierAdvanced},
					{"output_manager", "Output Manager", "Token budget, compression ratios", TierAdvanced},
				},
			},
			{
				Title: "AI & Knowledge",
				Categories: []Category{
					{"knowledge", "Knowledge", "Learning, Context limits", TierBasic},
					{"skill", "Skill", "File-based skill system", TierBasic},
					{"observational_memory", "Observational Memory", "Observer, Reflector, Thresholds", TierBasic},
					{"embedding", "Embedding & RAG", "Provider, Model, RAG settings", TierBasic},
					{"graph", "Graph Store", "Knowledge graph, GraphRAG settings", TierAdvanced},
					{"librarian", "Librarian", "Proactive knowledge extraction", TierAdvanced},
					{"agent_memory", "Agent Memory", "Per-agent persistent memory", TierAdvanced},
					{"multi_agent", "Multi-Agent", "Orchestration mode", TierAdvanced},
					{"a2a", "A2A Protocol", "Agent-to-Agent, remote agents", TierAdvanced},
					{"hooks", "Hooks", "Tool execution hooks, security filter", TierAdvanced},
				},
			},
			{
				Title: "Automation",
				Categories: []Category{
					{"cron", "Cron Scheduler", "Scheduled jobs, timezone, history", TierBasic},
					{"background", "Background Tasks", "Async tasks, concurrency limits", TierAdvanced},
					{"workflow", "Workflow Engine", "DAG workflows, timeouts, state", TierAdvanced},
				},
			},
			{
				Title: "Payment & Account",
				Categories: []Category{
					{"payment", "Payment", "Blockchain wallet, spending limits", TierBasic},
					{"smartaccount", "Smart Account", "ERC-7579 account, session keys, modules", TierAdvanced},
					{"smartaccount_session", "SA Session Keys", "Duration, gas limits, active keys", TierAdvanced},
					{"smartaccount_paymaster", "SA Paymaster", "Gasless USDC transactions", TierAdvanced},
					{"smartaccount_modules", "SA Modules", "Module contract addresses", TierAdvanced},
				},
			},
			{
				Title: "P2P & Economy",
				Categories: []Category{
					{"p2p", "P2P Network", "Peer-to-peer networking, discovery", TierAdvanced},
					{"p2p_workspace", "P2P Workspace", "Workspaces, git bundles", TierAdvanced},
					{"p2p_zkp", "P2P ZKP", "Zero-knowledge proof settings", TierAdvanced},
					{"p2p_pricing", "P2P Pricing", "Paid tool invocations", TierAdvanced},
					{"p2p_owner", "P2P Owner Protection", "Owner PII leak prevention", TierAdvanced},
					{"p2p_sandbox", "P2P Sandbox", "Tool isolation, container sandbox", TierAdvanced},
					{"economy", "Economy", "Budget, risk, pricing settings", TierAdvanced},
					{"economy_risk", "Economy Risk", "Trust-based risk assessment", TierAdvanced},
					{"economy_negotiation", "Economy Negotiation", "P2P price negotiation", TierAdvanced},
					{"economy_escrow", "Economy Escrow", "Milestone-based escrow", TierAdvanced},
					{"economy_escrow_onchain", "On-Chain Escrow", "Hub/Vault mode, contracts, settlement", TierAdvanced},
					{"economy_pricing", "Economy Pricing", "Dynamic pricing rules", TierAdvanced},
				},
			},
			{
				Title: "Integrations",
				Categories: []Category{
					{"mcp", "MCP Settings", "Global MCP server settings", TierBasic},
					{"mcp_servers", "MCP Server List", "Add, edit, remove MCP servers", TierAdvanced},
					{"observability", "Observability", "Token tracking, health, metrics", TierAdvanced},
				},
			},
			{
				Title: "Security",
				Categories: []Category{
					{"security", "Security", "PII, Approval, Encryption", TierBasic},
					{"auth", "Auth", "OIDC provider configuration", TierAdvanced},
					{"security_db", "Security DB Encryption", "SQLCipher database encryption", TierAdvanced},
					{"security_kms", "Security KMS", "Cloud KMS / HSM backends", TierAdvanced},
				},
			},
			{
				Title: "",
				Categories: []Category{
					{"save", "Save & Exit", "Save encrypted profile", TierBasic},
					{"cancel", "Cancel", "Exit without saving", TierBasic},
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
			case "down":
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
		case "tab":
			m.showAdvanced = !m.showAdvanced
			// Clamp cursor to visible items.
			items := m.selectableItems()
			if m.Cursor >= len(items) {
				m.Cursor = len(items) - 1
			}
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			return m, nil
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
// Supports smart filter prefixes: @basic, @advanced, @modified, @enabled.
// Search always covers all categories regardless of tier.
func (m *MenuModel) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		m.filtered = nil
		m.Cursor = 0
		return
	}

	all := m.allCategories()

	// Handle smart filter prefixes
	switch query {
	case "@basic":
		var results []Category
		for _, cat := range all {
			if cat.Tier == TierBasic {
				results = append(results, cat)
			}
		}
		m.filtered = results
		m.Cursor = 0
		return
	case "@advanced":
		var results []Category
		for _, cat := range all {
			if cat.Tier == TierAdvanced {
				results = append(results, cat)
			}
		}
		m.filtered = results
		m.Cursor = 0
		return
	case "@modified":
		if m.DirtyChecker != nil {
			var results []Category
			for _, cat := range all {
				if m.DirtyChecker(cat.ID) {
					results = append(results, cat)
				}
			}
			m.filtered = results
			m.Cursor = 0
			return
		}
	case "@enabled":
		if m.EnabledChecker != nil {
			var results []Category
			for _, cat := range all {
				if m.EnabledChecker(cat.ID) {
					results = append(results, cat)
				}
			}
			m.filtered = results
			m.Cursor = 0
			return
		}
	}

	var results []Category
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
		// Show filter hints when search input is empty
		if strings.TrimSpace(m.searchInput.Value()) == "" {
			filterHint := lipgloss.NewStyle().
				Foreground(tui.Dim).
				Italic(true).
				PaddingLeft(1)
			b.WriteString("\n")
			b.WriteString(filterHint.Render("@basic  @advanced  @enabled  @modified"))
		}
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
			tui.HelpEntry("\u2191\u2193", "Navigate"),
			tui.HelpEntry("Enter", "Select"),
			tui.HelpEntry("Esc", "Cancel"),
		))
	} else {
		tierLabel := "Show All"
		if m.showAdvanced {
			tierLabel = "Basic Only"
		}
		b.WriteString(tui.HelpBar(
			tui.HelpEntry("\u2191\u2193", "Navigate"),
			tui.HelpEntry("Enter", "Select"),
			tui.HelpEntry("/", "Search"),
			tui.HelpEntry("Tab", tierLabel),
			tui.HelpEntry("Esc", "Back"),
		))
	}

	return b.String()
}

func (m MenuModel) renderGroupedView(b *strings.Builder) {
	globalIdx := 0
	first := true
	for _, section := range m.Sections {
		// Filter categories by tier.
		var visible []Category
		for _, c := range section.Categories {
			if m.showAdvanced || c.Tier == TierBasic {
				visible = append(visible, c)
			}
		}
		if len(visible) == 0 {
			continue
		}

		// Section header
		if section.Title != "" {
			if !first {
				b.WriteString(tui.SeparatorLineStyle.Render("  " + strings.Repeat("\u2500", 38)))
				b.WriteString("\n")
			}
			b.WriteString(tui.SectionHeaderStyle.Render(section.Title))
			b.WriteString("\n")
		} else if !first {
			b.WriteString(tui.SeparatorLineStyle.Render("  " + strings.Repeat("\u2500", 38)))
			b.WriteString("\n")
		}
		first = false

		for _, cat := range visible {
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
		cursor = tui.CursorStyle.Render("\u25b8 ")
		titleStyle = titleStyle.Foreground(tui.Accent).Bold(true)
		descStyle = descStyle.Foreground(tui.Accent)
	}

	// Handle search highlighting
	title := cat.Title
	desc := cat.Desc

	// ADV badge for advanced categories
	badge := ""
	if cat.Tier == TierAdvanced {
		badge = " " + tui.BadgeAdvancedStyle.Render("ADV")
	}

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
		b.WriteString(badge)
	} else {
		b.WriteString(cursor)
		b.WriteString(titleStyle.Render(title))
		if desc != "" {
			b.WriteString(descStyle.Render(desc))
		}
		b.WriteString(badge)
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
