package settings

import (
	"fmt"
	"strconv"
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

// menuLevel tracks the current navigation depth.
type menuLevel int

const (
	levelSections   menuLevel = iota // Level 1: section list
	levelCategories                  // Level 2: categories within a section
)

// MenuModel manages the configuration menu.
type MenuModel struct {
	Sections     []Section
	Cursor       int
	Selected     string
	Width        int
	Height       int
	showAdvanced bool

	// Hierarchical navigation
	level            menuLevel
	activeSectionIdx int // index into Sections for Level 2
	sectionCursor    int // cursor position at Level 1 (restored on Esc from Level 2)

	// Search
	searching   bool
	searchInput textinput.Model
	filtered    []Category // filtered results (nil when not searching)

	// Checkers for smart filters
	DirtyChecker      func(string) bool // returns true if category config has been modified
	EnabledChecker    func(string) bool // returns true if category feature is enabled
	DependencyChecker func(string) int  // returns count of unmet required dependencies
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

// ShowAdvanced returns the current advanced mode state.
func (m MenuModel) ShowAdvanced() bool {
	return m.showAdvanced
}

// InCategoryLevel returns true when the menu is at Level 2 (categories within a section).
func (m MenuModel) InCategoryLevel() bool {
	return m.level == levelCategories
}

// ActiveSectionTitle returns the title of the currently active section (Level 2).
func (m MenuModel) ActiveSectionTitle() string {
	if m.activeSectionIdx >= 0 && m.activeSectionIdx < len(m.Sections) {
		return m.Sections[m.activeSectionIdx].Title
	}
	return ""
}

// selectableItems returns the list the cursor currently navigates.
func (m *MenuModel) selectableItems() []Category {
	if m.searching && m.filtered != nil {
		return m.filtered
	}
	if m.level == levelSections {
		return m.level1Items()
	}
	return m.activeSectionCategories()
}

// level1Items builds the item list for Level 1 (section list + save/cancel).
func (m *MenuModel) level1Items() []Category {
	var items []Category
	for i, s := range m.Sections {
		if s.Title == "" {
			items = append(items, s.Categories...)
		} else {
			items = append(items, Category{
				ID:    fmt.Sprintf("__section_%d", i),
				Title: s.Title,
				Desc:  fmt.Sprintf("%d settings", len(s.Categories)),
			})
		}
	}
	return items
}

// activeSectionCategories returns categories from the active section, filtered by tier.
func (m *MenuModel) activeSectionCategories() []Category {
	if m.activeSectionIdx < 0 || m.activeSectionIdx >= len(m.Sections) {
		return nil
	}
	section := m.Sections[m.activeSectionIdx]
	var vis []Category
	for _, c := range section.Categories {
		if m.showAdvanced || c.Tier == TierBasic {
			vis = append(vis, c)
		}
	}
	return vis
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
		showAdvanced:     true,
		level:            levelSections,
		activeSectionIdx: -1,
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
		case "esc":
			if m.level == levelCategories {
				m.level = levelSections
				m.Cursor = m.sectionCursor
				return m, nil
			}
		case "tab":
			if m.level == levelSections {
				return m, nil
			}
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
			if len(items) == 0 || m.Cursor >= len(items) {
				return m, nil
			}
			item := items[m.Cursor]
			if m.level == levelSections && strings.HasPrefix(item.ID, "__section_") {
				sIdx, _ := strconv.Atoi(strings.TrimPrefix(item.ID, "__section_"))
				if sIdx >= 0 && sIdx < len(m.Sections) {
					m.sectionCursor = m.Cursor
					m.activeSectionIdx = sIdx
					m.level = levelCategories
					m.Cursor = 0
				}
				return m, nil
			}
			m.Selected = item.ID
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
	case "@ready":
		if m.DependencyChecker != nil {
			var results []Category
			for _, cat := range all {
				if m.DependencyChecker(cat.ID) == 0 {
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
			b.WriteString(filterHint.Render("@basic  @advanced  @enabled  @modified  @ready"))
		}
	} else {
		hint := lipgloss.NewStyle().
			Foreground(tui.Dim).
			Italic(true).
			PaddingLeft(1)
		b.WriteString(hint.Render("/ Search..."))
	}
	b.WriteString("\n\n")

	// Section header with tab indicator (Level 2 only, outside search results)
	if m.level == levelCategories && (!m.searching || m.filtered == nil) {
		section := m.Sections[m.activeSectionIdx]
		headerStyle := lipgloss.NewStyle().Foreground(tui.Primary).Bold(true).PaddingLeft(2)
		b.WriteString(headerStyle.Render(section.Title))
		b.WriteString("  ")
		b.WriteString(m.renderTabIndicator())
		b.WriteString("\n\n")
	}

	// Menu body
	var body strings.Builder
	if m.searching && m.filtered != nil {
		m.renderFilteredView(&body)
	} else if m.level == levelCategories {
		m.renderCategoryDetailView(&body)
	} else {
		m.renderSectionListView(&body)
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
	} else if m.level == levelCategories {
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
	} else {
		b.WriteString(tui.HelpBar(
			tui.HelpEntry("\u2191\u2193", "Navigate"),
			tui.HelpEntry("Enter", "Select"),
			tui.HelpEntry("/", "Search"),
			tui.HelpEntry("Esc", "Back"),
		))
	}

	return b.String()
}

func (m MenuModel) renderSectionListView(b *strings.Builder) {
	items := m.level1Items()
	namedCount := 0
	for _, s := range m.Sections {
		if s.Title != "" {
			namedCount++
		}
	}
	for i, item := range items {
		if i == namedCount {
			b.WriteString(tui.SeparatorLineStyle.Render("  " + strings.Repeat("\u2500", 38)))
			b.WriteString("\n")
		}
		m.renderItem(b, item, i)
	}
}

func (m MenuModel) renderCategoryDetailView(b *strings.Builder) {
	cats := m.activeSectionCategories()
	if len(cats) == 0 {
		noResult := lipgloss.NewStyle().Foreground(tui.Muted).Italic(true)
		b.WriteString(noResult.Render("  No basic settings. Press Tab to show all."))
		b.WriteString("\n")
		return
	}
	for i, cat := range cats {
		m.renderItem(b, cat, i)
	}
}

func (m MenuModel) renderTabIndicator() string {
	activeStyle := lipgloss.NewStyle().Foreground(tui.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(tui.Dim)
	if m.showAdvanced {
		return inactiveStyle.Render("[Basic]") + " " + activeStyle.Render("[All]")
	}
	return activeStyle.Render("[Basic]") + " " + inactiveStyle.Render("[All]")
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

	// Dependency warning badge
	depBadge := ""
	if m.DependencyChecker != nil {
		if n := m.DependencyChecker(cat.ID); n > 0 {
			depBadge = " " + tui.BadgeDependencyStyle.Render(fmt.Sprintf("⚠ %d", n))
		}
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
		b.WriteString(depBadge)
	} else {
		b.WriteString(cursor)
		b.WriteString(titleStyle.Render(title))
		if desc != "" {
			b.WriteString(descStyle.Render(desc))
		}
		b.WriteString(badge)
		b.WriteString(depBadge)
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
