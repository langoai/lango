package pages

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
)

// newTestApprovalsPage creates an ApprovalsPage with fixed time and sets dimensions via Update.
func newTestApprovalsPage(history *approval.HistoryStore, grants *approval.GrantStore, width, height int) *ApprovalsPage {
	p := NewApprovalsPage(history, grants)
	p.nowFn = func() time.Time { return time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC) }
	updated, _ := p.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(*ApprovalsPage)
}

// sampleHistory returns a HistoryStore populated with 3 entries.
func sampleHistory() *approval.HistoryStore {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	store := approval.NewHistoryStore(100)
	store.Append(approval.HistoryEntry{
		Timestamp: now.Add(-5 * time.Minute),
		ToolName:  "exec",
		Summary:   "Execute: go test",
		Outcome:   "session",
		Provider:  "tui",
	})
	store.Append(approval.HistoryEntry{
		Timestamp: now.Add(-2 * time.Minute),
		ToolName:  "fs_write",
		Summary:   "Write to main.go",
		Outcome:   "approved",
		Provider:  "tui",
	})
	store.Append(approval.HistoryEntry{
		Timestamp: now.Add(-30 * time.Second),
		ToolName:  "fs_read",
		Summary:   "Read config.yaml",
		Outcome:   "bypass",
		Provider:  "auto",
	})
	return store
}

// sampleGrants returns a GrantStore populated with 2 grants.
func sampleGrants() *approval.GrantStore {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	gs := approval.NewGrantStore()
	gs.Grant("tty:default", "exec")
	gs.Grant("tty:default", "fs_write")
	// Override nowFn to set grant times in the past for display.
	// The grants were created at "now" by default, so relative time will show "0s ago".
	_ = now
	return gs
}

func TestApprovalsPage_Title(t *testing.T) {
	t.Parallel()

	p := NewApprovalsPage(nil, nil)
	assert.Equal(t, "Approvals", p.Title())
}

func TestApprovalsPage_NilStores(t *testing.T) {
	t.Parallel()

	p := newTestApprovalsPage(nil, nil, 80, 24)
	view := p.View()
	assert.Contains(t, view, "No approval history yet.")
}

func TestApprovalsPage_EmptyStores(t *testing.T) {
	t.Parallel()

	history := approval.NewHistoryStore(100)
	grants := approval.NewGrantStore()
	p := newTestApprovalsPage(history, grants, 80, 24)
	p.Activate()
	view := p.View()
	assert.Contains(t, view, "No approval history yet.")
}

func TestApprovalsPage_HistoryDisplay(t *testing.T) {
	t.Parallel()

	history := sampleHistory()
	p := newTestApprovalsPage(history, nil, 120, 24)
	p.Activate()
	view := p.View()

	// Should display tool names and outcomes.
	assert.Contains(t, view, "exec")
	assert.Contains(t, view, "fs_write")
	assert.Contains(t, view, "fs_read")
	assert.Contains(t, view, "approved")
	assert.Contains(t, view, "session")
	assert.Contains(t, view, "bypass")
	assert.Contains(t, view, "Approval History")
}

func TestApprovalsPage_GrantsDisplay(t *testing.T) {
	t.Parallel()

	grants := sampleGrants()
	p := newTestApprovalsPage(nil, grants, 120, 24)
	p.Activate()

	// Switch to grants section to verify display.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)
	view := p.View()

	assert.Contains(t, view, "tty:default")
	assert.Contains(t, view, "exec")
	assert.Contains(t, view, "fs_write")
	assert.Contains(t, view, "Active Grants")
}

func TestApprovalsPage_TabSwitchesSection(t *testing.T) {
	t.Parallel()

	history := sampleHistory()
	grants := sampleGrants()
	p := newTestApprovalsPage(history, grants, 120, 24)
	p.Activate()

	// Initial section is history.
	assert.Equal(t, 0, p.section)

	// Move history cursor down first.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 1, p.cursor)

	// Tab to grants.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 1, p.section)
	assert.Equal(t, 0, p.grantCursor, "grant cursor should start at 0")

	// History cursor should be preserved.
	assert.Equal(t, 1, p.cursor, "history cursor should be preserved when switching")

	// Tab back to history.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 0, p.section)
	assert.Equal(t, 1, p.cursor, "history cursor should be preserved after round-trip")
}

func TestApprovalsPage_RevokeGrant(t *testing.T) {
	t.Parallel()

	grants := approval.NewGrantStore()
	grants.Grant("session-1", "exec")
	grants.Grant("session-1", "fs_write")

	p := newTestApprovalsPage(nil, grants, 120, 24)
	p.Activate()

	// Switch to grants section.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)
	require.Equal(t, 1, p.section)
	require.Len(t, p.grantList, 2)

	// Press 'r' to revoke the first grant.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	p = updated.(*ApprovalsPage)

	// One grant should be revoked.
	assert.Len(t, p.grantList, 1)
}

func TestApprovalsPage_RevokeSessionGrants(t *testing.T) {
	t.Parallel()

	grants := approval.NewGrantStore()
	grants.Grant("session-1", "exec")
	grants.Grant("session-1", "fs_write")
	grants.Grant("session-2", "exec")

	p := newTestApprovalsPage(nil, grants, 120, 24)
	p.Activate()

	// Switch to grants section.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)
	require.Equal(t, 1, p.section)
	require.Len(t, p.grantList, 3)

	// Cursor is at 0 (session-1 / exec). Press 'R' to revoke all session-1 grants.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	p = updated.(*ApprovalsPage)

	// Only session-2's grant should remain.
	assert.Len(t, p.grantList, 1)
	assert.Equal(t, "session-2", p.grantList[0].SessionKey)
}

func TestApprovalsPage_CursorNavigation(t *testing.T) {
	t.Parallel()

	history := sampleHistory()
	p := newTestApprovalsPage(history, nil, 120, 24)
	p.Activate()

	// Cursor starts at 0.
	assert.Equal(t, 0, p.cursor)

	// Move down.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 1, p.cursor)

	// Move down again.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 2, p.cursor)

	// Move down beyond last entry — should clamp.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 2, p.cursor, "cursor should not exceed last index")

	// Move up.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 1, p.cursor)

	// Move up to 0.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 0, p.cursor)

	// Move up beyond first entry — should clamp.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 0, p.cursor, "cursor should not go below 0")
}

func TestApprovalsPage_CursorHighlight(t *testing.T) {
	t.Parallel()

	history := sampleHistory()
	p := newTestApprovalsPage(history, nil, 120, 24)
	p.Activate()

	view := p.View()
	assert.Contains(t, view, ">", "active cursor should show > prefix")

	// Move cursor to second row.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	view = p.View()

	lines := strings.Split(view, "\n")
	var firstLine, secondLine string
	for _, line := range lines {
		if strings.Contains(line, "fs_read") {
			firstLine = line // newest entry (index 0 when cursor is at 1)
		}
		if strings.Contains(line, "fs_write") {
			secondLine = line // second entry
		}
	}
	require.NotEmpty(t, firstLine, "should find fs_read line")
	require.NotEmpty(t, secondLine, "should find fs_write line")
	assert.NotContains(t, firstLine, ">", "first entry should not have cursor when cursor=1")
	assert.Contains(t, secondLine, ">", "second entry should have cursor")
}

func TestApprovalsPage_Activate(t *testing.T) {
	t.Parallel()

	history := sampleHistory()
	p := NewApprovalsPage(history, nil)

	cmd := p.Activate()
	assert.True(t, p.tickActive, "tickActive should be true after Activate")
	assert.NotNil(t, cmd, "Activate should return a tick command")
	assert.Len(t, p.histEntries, 3, "history entries should be populated after Activate")
}

func TestApprovalsPage_Deactivate(t *testing.T) {
	t.Parallel()

	history := sampleHistory()
	p := NewApprovalsPage(history, nil)
	p.Activate()
	assert.True(t, p.tickActive)

	p.Deactivate()
	assert.False(t, p.tickActive, "tickActive should be false after Deactivate")

	// A tick message after deactivation should not re-enable the tick.
	updated, cmd := p.Update(approvalTickMsg(time.Now()))
	p = updated.(*ApprovalsPage)
	assert.Nil(t, cmd, "tick command should not be returned when deactivated")
}

func TestApprovalsPage_ShortHelp_HistorySection(t *testing.T) {
	t.Parallel()

	p := NewApprovalsPage(nil, nil)
	p.section = 0

	bindings := p.ShortHelp()
	assert.Len(t, bindings, 3, "history section should have 3 bindings (tab, up, down)")
}

func TestApprovalsPage_ShortHelp_GrantsSection(t *testing.T) {
	t.Parallel()

	grants := approval.NewGrantStore()
	p := NewApprovalsPage(nil, grants)
	p.section = 1

	bindings := p.ShortHelp()
	assert.Len(t, bindings, 5, "grants section should have 5 bindings (tab, up, down, r, R)")
}

func TestApprovalsPage_GrantCursorIndependence(t *testing.T) {
	t.Parallel()

	grants := approval.NewGrantStore()
	grants.Grant("s1", "tool-a")
	grants.Grant("s1", "tool-b")
	grants.Grant("s1", "tool-c")

	p := newTestApprovalsPage(nil, grants, 120, 24)
	p.Activate()

	// Switch to grants section.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)

	// Move grant cursor down.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 2, p.grantCursor)

	// Switch to history section.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*ApprovalsPage)
	assert.Equal(t, 0, p.section)

	// Grant cursor should be preserved.
	assert.Equal(t, 2, p.grantCursor, "grant cursor should be preserved when switching to history")
}

func TestRelativeTime(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		give string
		t    time.Time
		want string
	}{
		{give: "seconds ago", t: now.Add(-30 * time.Second), want: "30s ago"},
		{give: "minutes ago", t: now.Add(-5 * time.Minute), want: "5m ago"},
		{give: "hours ago", t: now.Add(-3 * time.Hour), want: "3h ago"},
		{give: "days ago", t: now.Add(-48 * time.Hour), want: "2d ago"},
		{give: "just now", t: now, want: "0s ago"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := tui.RelativeTime(now, tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}
