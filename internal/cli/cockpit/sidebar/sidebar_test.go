package sidebar

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSidebarView_Visible(t *testing.T) {
	m := New()
	m.SetHeight(10)

	got := m.View()
	require.NotEmpty(t, got, "visible sidebar must produce output")

	// All menu labels should appear in the rendered output.
	for _, it := range m.items {
		assert.Contains(t, got, it.Label,
			"sidebar should contain menu label %q", it.Label)
	}
}

func TestSidebarView_Hidden(t *testing.T) {
	m := New()
	m.SetVisible(false)

	got := m.View()
	assert.Empty(t, got, "hidden sidebar must return empty string")
}

func TestSidebarHeight(t *testing.T) {
	tests := []struct {
		give     int
		wantMin  int
		wantDesc string
	}{
		{
			give:     5,
			wantMin:  5,
			wantDesc: "height equal to item count",
		},
		{
			give:     20,
			wantMin:  20,
			wantDesc: "height larger than item count fills extra lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantDesc, func(t *testing.T) {
			m := New()
			m.SetHeight(tt.give)
			out := m.View()
			lines := strings.Split(out, "\n")
			assert.GreaterOrEqual(t, len(lines), tt.wantMin,
				"rendered lines should be >= requested height %d", tt.give)
		})
	}
}

func TestSidebarNonInteractive(t *testing.T) {
	m := New()
	m.SetHeight(10)

	beforeView := m.View()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Nil(t, cmd, "Update should return nil cmd")

	// The returned model should produce the same view.
	afterView := updated.(Model).View()
	assert.Equal(t, beforeView, afterView,
		"Update must not change sidebar state")
}

// --- Mouse click tests ---

func TestMouseClick_SelectsEnabledItem(t *testing.T) {
	m := New()
	m.SetHeight(10)

	// Click on item index 1 ("settings") — Y=1 is the second row.
	msg := tea.MouseMsg{
		X:      5,
		Y:      1,
		Action: tea.MouseActionRelease,
	}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "click on enabled item should produce a command")

	result := cmd()
	psm, ok := result.(PageSelectedMsg)
	require.True(t, ok, "command should produce PageSelectedMsg")
	assert.Equal(t, "settings", psm.ID)
}

func TestMouseClick_IgnoresDisabledItem(t *testing.T) {
	m := New()
	m.SetHeight(10)
	// Manually disable an item for the test.
	m.items[4].Disabled = true

	msg := tea.MouseMsg{
		X:      5,
		Y:      4,
		Action: tea.MouseActionRelease,
	}
	_, cmd := m.Update(msg)
	assert.Nil(t, cmd, "click on disabled item should not produce a command")
}

func TestMouseClick_IgnoresOutOfBounds(t *testing.T) {
	m := New()
	m.SetHeight(10)

	// Y=99 is beyond the item list.
	msg := tea.MouseMsg{
		X:      5,
		Y:      99,
		Action: tea.MouseActionRelease,
	}
	_, cmd := m.Update(msg)
	assert.Nil(t, cmd, "click out of bounds should not produce a command")
}

func TestMouseClick_WorksWhenUnfocused(t *testing.T) {
	m := New()
	m.SetHeight(10)
	m.SetFocused(false) // explicitly unfocused

	msg := tea.MouseMsg{
		X:      5,
		Y:      2, // "tools"
		Action: tea.MouseActionRelease,
	}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "mouse click should work even when sidebar is unfocused")

	result := cmd()
	psm, ok := result.(PageSelectedMsg)
	require.True(t, ok)
	assert.Equal(t, "tools", psm.ID)
}

func TestMouseClick_IgnoresNonRelease(t *testing.T) {
	m := New()
	m.SetHeight(10)

	msg := tea.MouseMsg{
		X:      5,
		Y:      1,
		Action: tea.MouseActionPress,
	}
	_, cmd := m.Update(msg)
	assert.Nil(t, cmd, "non-release mouse action should not produce a command")
}
