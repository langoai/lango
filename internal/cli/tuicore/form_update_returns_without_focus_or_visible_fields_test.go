package tuicore

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormInitReturnsBlinkCommand(t *testing.T) {
	form := NewFormModel("Fixture Form")

	cmd := form.Init()

	require.NotNil(t, cmd)
}

func TestAddFieldInitializesPasswordAndSearchSelectState(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.AddField(&Field{
		Key:         "password",
		Label:       "Password",
		Type:        InputPassword,
		Value:       "secret",
		Placeholder: "enter password",
		Width:       12,
	})

	password := form.Fields[0]
	assert.Equal(t, "secret", password.InitialValue)
	assert.Equal(t, "secret", password.TextInput.Value())
	assert.Equal(t, "enter password", password.TextInput.Placeholder)
	assert.Equal(t, 12, password.TextInput.Width)
	assert.Equal(t, textinput.EchoPassword, password.TextInput.EchoMode)
	assert.Equal(t, '*', password.TextInput.EchoCharacter)

	options := []string{"alpha", "beta"}
	form.AddField(&Field{
		Key:     "model",
		Label:   "Model",
		Type:    InputSearchSelect,
		Value:   "alpha",
		Options: options,
		Width:   22,
	})

	options[0] = "mutated"
	search := form.Fields[1]
	assert.Equal(t, "alpha", search.InitialValue)
	assert.Equal(t, "alpha", search.TextInput.Value())
	assert.Equal(t, "Type to search...", search.TextInput.Placeholder)
	assert.Equal(t, 22, search.TextInput.Width)
	assert.Equal(t, []string{"alpha", "beta"}, search.FilteredOptions)
}

func TestFormUpdateReturnsWithoutFocusOrVisibleFields(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = false
	form.AddField(&Field{
		Key:   "hidden",
		Label: "Hidden",
		Type:  InputText,
		Value: "unchanged",
	})

	updated, cmd := form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	require.Nil(t, cmd)
	assert.Equal(t, "unchanged", updated.Fields[0].Value)

	form.Focus = true
	form.Fields[0].VisibleWhen = func() bool { return false }
	form.Cursor = 12

	updated, cmd = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Nil(t, cmd)
	assert.Equal(t, 12, updated.Cursor)
	assert.Empty(t, updated.VisibleFields())
}

func TestFormFieldOptionsLoadedFallsBackToManualInputForEmptyOptions(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.AddField(&Field{
		Key:         "model",
		Label:       "Model",
		Type:        InputSearchSelect,
		Value:       "manual-model",
		Placeholder: "model id",
		Options:     []string{"old-model"},
		Loading:     true,
		Width:       17,
	})

	updated, cmd := form.Update(FieldOptionsLoadedMsg{
		FieldKey:   "model",
		ProviderID: "provider-a",
		Options:    nil,
	})

	require.Nil(t, cmd)
	field := updated.Fields[0]
	assert.False(t, field.Loading)
	assert.NoError(t, field.LoadError)
	assert.Equal(t, InputText, field.Type)
	assert.Equal(t, "manual-model", field.TextInput.Value())
	assert.Equal(t, 17, field.TextInput.Width)
}

func TestSearchSelectDropdownFiltersSelectsAndMarksEdited(t *testing.T) {
	form := newTestSearchSelectForm([]string{"alpha", "beta", "betamax", "gamma"}, "alpha")
	field := form.Fields[0]

	var cmd tea.Cmd
	form, cmd = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)
	require.True(t, field.SelectOpen)

	form, cmd = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bet")})
	require.NotNil(t, cmd)
	assert.Equal(t, []string{"beta", "betamax"}, field.FilteredOptions)
	assert.Equal(t, 0, field.SelectCursor)

	form, cmd = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Nil(t, cmd)
	assert.Equal(t, 1, field.SelectCursor)

	form, cmd = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)
	assert.False(t, field.SelectOpen)
	assert.Equal(t, "betamax", field.Value)
	assert.Equal(t, "betamax", field.TextInput.Value())
	assert.True(t, field.Edited)
}

func TestFormCursorClampsAfterVisibilityChanges(t *testing.T) {
	showSecond := true
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.AddField(&Field{
		Key:   "toggle",
		Label: "Toggle",
		Type:  InputBool,
	})
	form.AddField(&Field{
		Key:         "details",
		Label:       "Details",
		Type:        InputText,
		Value:       "visible",
		VisibleWhen: func() bool { return showSecond },
	})
	form.Cursor = 1

	showSecond = false
	updated, cmd := form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	require.Nil(t, cmd)
	assert.Len(t, updated.VisibleFields(), 1)
	assert.Equal(t, 0, updated.Cursor)
	assert.Equal(t, "visible", updated.Fields[1].Value)
}

func TestFormUpdateHandlesNavigationTextToggleAndCancelBranches(t *testing.T) {
	cancelled := false
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.OnCancel = func() { cancelled = true }
	form.AddField(&Field{
		Key:     "enabled",
		Label:   "Enabled",
		Type:    InputBool,
		Checked: false,
	})
	form.AddField(&Field{
		Key:   "name",
		Label: "Name",
		Type:  InputText,
		Value: "bot",
	})

	updated, cmd := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Nil(t, cmd)
	assert.Equal(t, 1, updated.Cursor)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	require.Nil(t, cmd)
	assert.Equal(t, 0, updated.Cursor)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	require.Nil(t, cmd)
	assert.True(t, updated.Fields[0].Checked)
	assert.True(t, updated.Fields[0].Edited)

	updated.Cursor = 1
	updated.Fields[1].TextInput.Focus()
	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	assert.Equal(t, "botx", updated.Fields[1].Value)
	assert.True(t, updated.Fields[1].Edited)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Nil(t, cmd)
	assert.True(t, cancelled)
}

func TestFormUpdateReturnsWhenToggleHidesAllFields(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = true
	toggle := &Field{
		Key:   "toggle",
		Label: "Toggle",
		Type:  InputBool,
	}
	toggle.VisibleWhen = func() bool { return !toggle.Checked }
	form.AddField(toggle)

	updated, cmd := form.Update(tea.KeyMsg{Type: tea.KeySpace})

	require.Nil(t, cmd)
	assert.True(t, updated.Fields[0].Checked)
	assert.Empty(t, updated.VisibleFields())
}

func TestInputSelectUnknownValueCyclesAndRunsReturnedCommand(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.AddField(&Field{
		Key:     "provider",
		Label:   "Provider",
		Type:    InputSelect,
		Value:   "unknown",
		Options: []string{"anthropic", "openai"},
		OnChange: func(newValue string) tea.Cmd {
			return func() tea.Msg { return "changed:" + newValue }
		},
	})

	updated, cmd := form.Update(tea.KeyMsg{Type: tea.KeyRight})

	require.NotNil(t, cmd)
	assert.Equal(t, "anthropic", updated.Fields[0].Value)
	assert.True(t, updated.Fields[0].Edited)
	assert.Equal(t, "changed:anthropic", cmd())

	updated.Fields[0].Value = "missing"
	updated.Fields[0].Edited = false
	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyLeft})

	require.NotNil(t, cmd)
	assert.Equal(t, "openai", updated.Fields[0].Value)
	assert.True(t, updated.Fields[0].Edited)
	assert.Equal(t, "changed:openai", cmd())
}

func TestFormViewRendersLoadingReadOnlyAndDropdownHelpBranches(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.AddField(&Field{
		Key:     "loading",
		Label:   "Loading",
		Type:    InputSearchSelect,
		Loading: true,
	})
	form.AddField(&Field{
		Key:         "readonly",
		Label:       "Read Only",
		Type:        InputReadOnly,
		Value:       "fixed",
		Description: "Displayed for context",
	})
	form.AddField(&Field{
		Key:             "models",
		Label:           "Models",
		Type:            InputSearchSelect,
		Value:           "model-09",
		Options:         []string{"model-01", "model-02", "model-03", "model-04", "model-05", "model-06", "model-07", "model-08", "model-09", "model-10"},
		FilteredOptions: []string{"model-01", "model-02", "model-03", "model-04", "model-05", "model-06", "model-07", "model-08", "model-09", "model-10"},
		SelectOpen:      true,
		SelectCursor:    9,
	})

	form.Cursor = 0
	loadingView := form.View()
	assert.Contains(t, loadingView, "Loading models...")

	form.Fields[2].SelectOpen = false
	form.Cursor = 1
	readOnlyView := form.View()
	assert.Contains(t, readOnlyView, "fixed")
	assert.Contains(t, readOnlyView, "Read-only")
	assert.Contains(t, readOnlyView, "Displayed for context")

	form.Fields[2].SelectOpen = true
	form.Cursor = 2
	dropdownView := form.View()
	assert.Contains(t, dropdownView, "10/10 matches")
	assert.Contains(t, dropdownView, "model-10")
	assert.Contains(t, dropdownView, "Filter")
	assert.False(t, strings.Contains(dropdownView, "Read-only"))
}

func TestFormViewRendersEditableControlsAndDefaultHelpBranches(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.AddField(&Field{
		Key:   "name",
		Label: "Name",
		Type:  InputText,
		Value: "bot",
	})
	form.AddField(&Field{
		Key:     "enabled",
		Label:   "Enabled",
		Type:    InputBool,
		Checked: true,
	})
	form.AddField(&Field{
		Key:     "provider",
		Label:   "Provider",
		Type:    InputSelect,
		Options: []string{"openai", "anthropic"},
	})
	form.AddField(&Field{
		Key:   "model",
		Label: "Model",
		Type:  InputSearchSelect,
	})

	textView := form.View()
	assert.Contains(t, textView, "bot")
	assert.Contains(t, textView, "[x]")
	assert.Contains(t, textView, "openai")
	assert.Contains(t, textView, "(none)")
	assert.Contains(t, textView, "Toggle")
	assert.Contains(t, textView, "Search")

	form.Cursor = 2
	selectView := form.View()
	assert.Contains(t, selectView, "< openai >")

	form.Cursor = 3
	searchView := form.View()
	assert.Contains(t, searchView, "(none)  [Enter: search]")
}

func TestFormViewRendersDropdownOverflowIndicator(t *testing.T) {
	form := NewFormModel("Fixture Form")
	form.Focus = true
	form.AddField(&Field{
		Key:             "models",
		Label:           "Models",
		Type:            InputSearchSelect,
		Options:         []string{"model-01", "model-02", "model-03", "model-04", "model-05", "model-06", "model-07", "model-08", "model-09", "model-10"},
		FilteredOptions: []string{"model-01", "model-02", "model-03", "model-04", "model-05", "model-06", "model-07", "model-08", "model-09", "model-10"},
		SelectOpen:      true,
	})

	view := form.View()

	assert.Contains(t, view, "10/10 matches")
	assert.Contains(t, view, "... 2 more")
}
