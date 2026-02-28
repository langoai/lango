package tuicore

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSearchSelectForm(options []string, value string) FormModel {
	form := NewFormModel("Test Form")
	form.Focus = true
	form.AddField(&Field{
		Key:     "model",
		Label:   "Model",
		Type:    InputSearchSelect,
		Value:   value,
		Options: options,
	})
	return form
}

func TestInputSearchSelect_FilterBySubstring(t *testing.T) {
	tests := []struct {
		give       string
		wantCount  int
		wantFirst  string
	}{
		{give: "", wantCount: 4, wantFirst: "claude-3-opus"},
		{give: "claude", wantCount: 2, wantFirst: "claude-3-opus"},
		{give: "gpt", wantCount: 1, wantFirst: "gpt-4o"},
		{give: "xyz", wantCount: 0},
		{give: "CLAUDE", wantCount: 2, wantFirst: "claude-3-opus"}, // case insensitive
	}

	options := []string{"claude-3-opus", "claude-3-sonnet", "gpt-4o", "gemini-pro"}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			form := newTestSearchSelectForm(options, "")
			field := form.Fields[0]

			field.applySearchFilter(tt.give)

			assert.Equal(t, tt.wantCount, len(field.FilteredOptions))
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, field.FilteredOptions[0])
			}
		})
	}
}

func TestInputSearchSelect_OpenCloseWithEnterEsc(t *testing.T) {
	form := newTestSearchSelectForm([]string{"model-a", "model-b", "model-c"}, "model-b")
	field := form.Fields[0]

	// Initially closed
	assert.False(t, field.SelectOpen)

	// Press Enter to open dropdown
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, field.SelectOpen)

	// Cursor should be at current value
	assert.Equal(t, 1, field.SelectCursor) // model-b is index 1

	// Press Esc to close dropdown
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, field.SelectOpen)

	// Value should remain unchanged
	assert.Equal(t, "model-b", field.Value)
}

func TestInputSearchSelect_NavigateAndSelect(t *testing.T) {
	form := newTestSearchSelectForm([]string{"alpha", "beta", "gamma"}, "alpha")
	field := form.Fields[0]

	// Open dropdown
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, field.SelectOpen)
	assert.Equal(t, 0, field.SelectCursor) // alpha at index 0

	// Navigate down
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, field.SelectCursor)

	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, field.SelectCursor)

	// Don't go past the end
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, field.SelectCursor)

	// Navigate up
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, field.SelectCursor)

	// Select with Enter
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, field.SelectOpen)
	assert.Equal(t, "beta", field.Value)
}

func TestInputSearchSelect_TabClosesDropdown(t *testing.T) {
	form := NewFormModel("Test")
	form.Focus = true
	form.AddField(&Field{
		Key: "model", Label: "Model", Type: InputSearchSelect,
		Value:   "a",
		Options: []string{"a", "b", "c"},
	})
	form.AddField(&Field{
		Key: "name", Label: "Name", Type: InputText,
		Value: "test",
	})

	field := form.Fields[0]

	// Open dropdown
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, field.SelectOpen)

	// Tab should close dropdown and move to next field
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.False(t, field.SelectOpen)
	assert.Equal(t, 1, form.Cursor)
}

func TestInputSearchSelect_EscDoesNotCancelForm(t *testing.T) {
	cancelled := false
	form := newTestSearchSelectForm([]string{"a", "b"}, "a")
	form.OnCancel = func() { cancelled = true }
	field := form.Fields[0]

	// Open dropdown
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, field.SelectOpen)

	// Esc should close dropdown, NOT cancel form
	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, field.SelectOpen)
	assert.False(t, cancelled)
}

func TestInputSearchSelect_CursorClamping(t *testing.T) {
	form := newTestSearchSelectForm([]string{"embed-a", "embed-b", "gpt-4o"}, "")
	field := form.Fields[0]

	field.SelectCursor = 2
	field.applySearchFilter("embed")

	// Cursor should be clamped to new filtered length
	assert.Equal(t, 1, field.SelectCursor) // max index is 1 (2 items)
}

func TestFormModel_HasOpenDropdown(t *testing.T) {
	form := NewFormModel("Test")
	form.Focus = true
	form.AddField(&Field{
		Key: "model", Label: "Model", Type: InputSearchSelect,
		Options: []string{"a", "b"},
	})
	form.AddField(&Field{
		Key: "name", Label: "Name", Type: InputText,
	})

	assert.False(t, form.HasOpenDropdown())

	form.Fields[0].SelectOpen = true
	assert.True(t, form.HasOpenDropdown())
}
