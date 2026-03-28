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
