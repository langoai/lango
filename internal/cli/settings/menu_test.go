package settings

import (
	"sort"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExperimentalCategories_MatchesExpected verifies that the
// ExperimentalCategories map contains exactly the intended set.
// If a category is added or removed, this test will fail —
// prompting a deliberate update to the map.
func TestExperimentalCategories_MatchesExpected(t *testing.T) {
	t.Parallel()

	expected := []string{
		"a2a",
		"agent_memory",
		"alerting",
		"economy",
		"economy_escrow",
		"economy_escrow_onchain",
		"economy_negotiation",
		"economy_pricing",
		"economy_risk",
		"graph",
		"hooks",
		"librarian",
		"multi_agent",
		"observability",
		"ontology",
		"os_sandbox",
		"p2p",
		"p2p_owner",
		"p2p_pricing",
		"p2p_sandbox",
		"p2p_workspace",
		"p2p_zkp",
		"provenance",
		"runledger",
		"smartaccount",
		"smartaccount_modules",
		"smartaccount_paymaster",
		"smartaccount_session",
	}

	var got []string
	for id := range ExperimentalCategories {
		got = append(got, id)
	}
	sort.Strings(got)

	assert.Equal(t, expected, got,
		"ExperimentalCategories drift detected — update the map or this test")
}

func TestMenu_ApplyFilterSmartPrefixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		configure  func(*MenuModel)
		wantIDs    []string
		wantAbsent []string
	}{
		{
			name:  "basic",
			query: "@basic",
			wantIDs: []string{
				"providers",
				"agent",
				"knowledge",
				"payment",
				"mcp",
				"security",
				"save",
				"cancel",
			},
			wantAbsent: []string{"server", "graph", "p2p"},
		},
		{
			name:  "advanced",
			query: "@advanced",
			wantIDs: []string{
				"server",
				"graph",
				"p2p",
				"economy_escrow",
				"observability",
			},
			wantAbsent: []string{"providers", "agent", "save"},
		},
		{
			name:  "modified",
			query: "@modified",
			configure: func(m *MenuModel) {
				dirty := map[string]bool{"agent": true, "mcp": true}
				m.DirtyChecker = func(id string) bool {
					return dirty[id]
				}
			},
			wantIDs:    []string{"agent", "mcp"},
			wantAbsent: []string{"providers", "security"},
		},
		{
			name:  "enabled",
			query: "@enabled",
			configure: func(m *MenuModel) {
				enabled := map[string]bool{"knowledge": true, "security": true}
				m.EnabledChecker = func(id string) bool {
					return enabled[id]
				}
			},
			wantIDs:    []string{"knowledge", "security"},
			wantAbsent: []string{"agent", "mcp"},
		},
		{
			name:  "ready",
			query: "@ready",
			configure: func(m *MenuModel) {
				blocked := map[string]int{"p2p": 2, "economy": 1}
				m.DependencyChecker = func(id string) int {
					return blocked[id]
				}
			},
			wantIDs:    []string{"agent", "knowledge", "save"},
			wantAbsent: []string{"p2p", "economy"},
		},
		{
			name:       "experimental",
			query:      "@experimental",
			wantIDs:    []string{"p2p", "economy", "ontology", "runledger"},
			wantAbsent: []string{"agent", "knowledge", "save"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewMenuModel()
			if tt.configure != nil {
				tt.configure(&m)
			}
			m.searching = true
			m.searchInput.SetValue(tt.query)
			m.applyFilter()

			got := categoryIDs(m.filtered)
			for _, want := range tt.wantIDs {
				assert.Contains(t, got, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, got, absent)
			}
			assert.Equal(t, 0, m.Cursor, "filtering should reset navigation to the first result")
		})
	}
}

func TestMenu_SearchSelectClearsSearchState(t *testing.T) {
	m := NewMenuModel()
	m.searching = true
	m.searchInput.SetValue("wallet")
	m.applyFilter()
	require.NotEmpty(t, m.filtered)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, "payment", m.Selected)
	assert.False(t, m.IsSearching())
	assert.Empty(t, m.searchInput.Value())
	assert.Nil(t, m.filtered)
}

func TestMenu_ViewRendersSearchAndTierEmptyStates(t *testing.T) {
	m := NewMenuModel()
	m.searching = true
	m.searchInput.SetValue("no-such-settings-category")
	m.filtered = []Category{}

	assert.Contains(t, m.View(), "No matching items")

	m = NewMenuModel()
	m.activeSectionIdx = 4 // P2P & Economy contains advanced-only categories.
	m.level = levelCategories
	m.showAdvanced = false

	view := m.View()
	assert.Contains(t, view, "No basic settings")
	assert.Contains(t, view, "Show All")
}

func TestMenu_ViewRendersBadgesAndHighlighting(t *testing.T) {
	m := NewMenuModel()
	m.searching = true
	m.searchInput.SetValue("p2p")
	m.applyFilter()
	m.DependencyChecker = func(id string) int {
		if id == "p2p" {
			return 2
		}
		return 0
	}

	view := m.View()

	assert.Contains(t, view, "P2P")
	assert.Contains(t, view, "ADV")
	assert.Contains(t, view, "EXP")
	assert.Contains(t, view, "2")
	assert.True(t, strings.Contains(m.highlightMatch("P2P Network", "p2p", true), "P2P"))
}

func categoryIDs(categories []Category) []string {
	ids := make([]string, 0, len(categories))
	for _, cat := range categories {
		ids = append(ids, cat.ID)
	}
	return ids
}
