package pages

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolcatalog"
)

// newTestCatalog creates a Catalog pre-populated with two categories.
func newTestCatalog() *toolcatalog.Catalog {
	c := toolcatalog.New()
	c.RegisterCategory(toolcatalog.Category{
		Name:        "exec",
		Description: "Execution tools",
		Enabled:     true,
	})
	c.RegisterCategory(toolcatalog.Category{
		Name:        "browser",
		Description: "Browser tools",
		Enabled:     true,
	})
	c.Register("exec", []*agent.Tool{
		{Name: "exec_command", Description: "Execute shell commands", SafetyLevel: agent.SafetyLevelDangerous},
		{Name: "exec_run", Description: "Run a process", SafetyLevel: agent.SafetyLevelModerate},
	})
	c.Register("browser", []*agent.Tool{
		{Name: "browser_navigate", Description: "Navigate to URL", SafetyLevel: agent.SafetyLevelSafe},
	})
	return c
}

func TestToolsPage_Title(t *testing.T) {
	t.Parallel()

	p := NewToolsPage(toolcatalog.New())
	assert.Equal(t, "Tools", p.Title())
}

func TestToolsPage_ShortHelp(t *testing.T) {
	t.Parallel()

	p := NewToolsPage(toolcatalog.New())
	bindings := p.ShortHelp()
	assert.Len(t, bindings, 3, "expected up, down, back bindings")
}

func TestToolsPage_Categories(t *testing.T) {
	t.Parallel()

	cat := newTestCatalog()
	p := NewToolsPage(cat)

	// Categories are sorted alphabetically by the catalog.
	require.Len(t, p.categories, 2)
	assert.Equal(t, "browser", p.categories[0].Name)
	assert.Equal(t, "exec", p.categories[1].Name)

	// Initial cursor is 0 → browser category selected.
	assert.Equal(t, 0, p.categoryCursor)
	require.Len(t, p.tools, 1)
	assert.Equal(t, "browser_navigate", p.tools[0].Name)
}

func TestToolsPage_CursorNavigation(t *testing.T) {
	t.Parallel()

	cat := newTestCatalog()
	p := NewToolsPage(cat)

	// Move down: browser → exec.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ToolsPage)
	assert.Equal(t, 1, p.categoryCursor)
	require.Len(t, p.tools, 2)
	assert.Equal(t, "exec_command", p.tools[0].Name)

	// Move down again: should clamp at last index.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ToolsPage)
	assert.Equal(t, 1, p.categoryCursor, "cursor should not exceed last index")

	// Move up: exec → browser.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*ToolsPage)
	assert.Equal(t, 0, p.categoryCursor)
	require.Len(t, p.tools, 1)
	assert.Equal(t, "browser_navigate", p.tools[0].Name)

	// Move up again: should clamp at 0.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*ToolsPage)
	assert.Equal(t, 0, p.categoryCursor, "cursor should not go below 0")
}

func TestToolsPage_EmptyCatalog(t *testing.T) {
	t.Parallel()

	p := NewToolsPage(toolcatalog.New())
	assert.Empty(t, p.categories)
	assert.Empty(t, p.tools)

	// View should not panic with zero size.
	assert.Equal(t, "", p.View())

	// View should not panic with a valid size but no data.
	updated, _ := p.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	p = updated.(*ToolsPage)
	view := p.View()
	assert.Contains(t, view, "No categories")
}

func TestToolsPage_WindowSize(t *testing.T) {
	t.Parallel()

	p := NewToolsPage(newTestCatalog())
	updated, _ := p.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	p = updated.(*ToolsPage)

	assert.Equal(t, 120, p.width)
	assert.Equal(t, 40, p.height)
}

func TestToolsPage_Activate(t *testing.T) {
	t.Parallel()

	cat := newTestCatalog()
	p := NewToolsPage(cat)

	// Add a new category after construction.
	cat.RegisterCategory(toolcatalog.Category{
		Name:        "crypto",
		Description: "Crypto tools",
		Enabled:     true,
	})
	cat.Register("crypto", []*agent.Tool{
		{Name: "sign_message", Description: "Sign a message", SafetyLevel: agent.SafetyLevelModerate},
	})

	// Activate should refresh and pick up the new category.
	cmd := p.Activate()
	assert.Nil(t, cmd)
	assert.Len(t, p.categories, 3)
}

func TestToolsPage_ViewRendersContent(t *testing.T) {
	t.Parallel()

	p := NewToolsPage(newTestCatalog())
	updated, _ := p.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	p = updated.(*ToolsPage)

	view := p.View()
	assert.Contains(t, view, "CATEGORIES")
	assert.Contains(t, view, "browser")
	assert.Contains(t, view, "exec")
	assert.Contains(t, view, "browser_navigate")
}

func TestToolsPage_SafetyLevelColors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
	}{
		{give: "safe"},
		{give: "moderate"},
		{give: "dangerous"},
		{give: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			// safetyStyle should not panic for any input.
			s := safetyStyle(tt.give)
			result := s.Render(tt.give)
			assert.NotEmpty(t, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		giveMax int
		want    string
	}{
		{give: "short", giveMax: 10, want: "short"},
		{give: "exactly10!", giveMax: 10, want: "exactly10!"},
		{give: "this is too long", giveMax: 10, want: "this is..."},
		{give: "ab", giveMax: 2, want: "ab"},
		{give: "abc", giveMax: 2, want: "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := truncate(tt.give, tt.giveMax)
			assert.Equal(t, tt.want, got)
		})
	}
}
