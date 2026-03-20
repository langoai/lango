package settings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditor_EscAtWelcome_Quits(t *testing.T) {
	e := NewEditor()
	require.Equal(t, StepWelcome, e.step)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed := model.(*Editor)

	assert.Equal(t, StepWelcome, ed.step)
	assert.NotNil(t, cmd, "esc at welcome should return quit cmd")
}

func TestEditor_EscAtMenu_NavigatesToWelcome(t *testing.T) {
	e := NewEditor()
	e.step = StepMenu

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed := model.(*Editor)

	assert.Equal(t, StepWelcome, ed.step)
	assert.Nil(t, cmd, "esc at menu should not quit, just navigate back")
}

func TestEditor_EscAtMenuWhileSearching_StaysAtMenu(t *testing.T) {
	e := NewEditor()
	e.step = StepMenu

	// Enter search mode by pressing /
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	ed := model.(*Editor)
	require.True(t, ed.menu.IsSearching(), "should be in search mode")

	// Press esc — should cancel search, not navigate back
	model, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed = model.(*Editor)

	assert.Equal(t, StepMenu, ed.step, "should stay at menu")
	assert.False(t, ed.menu.IsSearching(), "search should be cancelled")
	assert.Nil(t, cmd)
}

func TestEditor_EscAtMenuLevel2_StaysAtMenu(t *testing.T) {
	e := NewEditor()
	e.step = StepMenu

	// Enter section (cursor at 0 = Core)
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ed := model.(*Editor)
	require.Equal(t, StepMenu, ed.step, "should still be at menu")
	require.True(t, ed.menu.InCategoryLevel(), "should be at category level")

	// Press Esc — should go back to section level, NOT to Welcome
	model, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed = model.(*Editor)

	assert.Equal(t, StepMenu, ed.step, "should stay at menu")
	assert.False(t, ed.menu.InCategoryLevel(), "should be back at section level")
	assert.Nil(t, cmd)
}

func TestMenu_EnterSection_TransitionsToLevel2(t *testing.T) {
	m := NewMenuModel()

	// Cursor at 0 = Core section
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, m.InCategoryLevel(), "should be at category level")
	assert.Equal(t, "Core", m.ActiveSectionTitle())
	assert.Equal(t, 0, m.Cursor, "cursor should reset to 0")
}

func TestMenu_EscAtLevel2_ReturnsToLevel1(t *testing.T) {
	m := NewMenuModel()

	// Navigate to section 2 (Automation)
	m.Cursor = 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, m.InCategoryLevel())
	require.Equal(t, "Automation", m.ActiveSectionTitle())

	// Press Esc
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, m.InCategoryLevel(), "should be back at section level")
	assert.Equal(t, 2, m.Cursor, "cursor should restore to section position")
}

func TestMenu_TabOnlyAtLevel2(t *testing.T) {
	m := NewMenuModel()

	// Tab at Level 1: should be no-op (showAdvanced stays true)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, m.ShowAdvanced(), "tab at Level 1 should be no-op")
	assert.False(t, m.InCategoryLevel())

	// Enter section to go to Level 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, m.InCategoryLevel())

	// Tab at Level 2: should toggle
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.False(t, m.ShowAdvanced(), "tab at Level 2 should toggle to basic")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, m.ShowAdvanced(), "tab again should toggle back")
}

func TestMenu_SearchAtBothLevels(t *testing.T) {
	m := NewMenuModel()

	// Search from Level 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, m.IsSearching(), "should enter search from Level 1")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m.IsSearching())

	// Enter Level 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, m.InCategoryLevel())

	// Search from Level 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, m.IsSearching(), "should enter search from Level 2")
}

func TestMenu_SaveCancelFromLevel1(t *testing.T) {
	m := NewMenuModel()

	// Navigate to Save & Exit (after 7 named sections = index 7)
	items := m.selectableItems()
	saveIdx := -1
	for i, item := range items {
		if item.ID == "save" {
			saveIdx = i
			break
		}
	}
	require.NotEqual(t, -1, saveIdx, "should find save item")

	m.Cursor = saveIdx
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "save", m.Selected, "should select save")

	// Reset and test cancel
	m = NewMenuModel()
	cancelIdx := -1
	items = m.selectableItems()
	for i, item := range items {
		if item.ID == "cancel" {
			cancelIdx = i
			break
		}
	}
	require.NotEqual(t, -1, cancelIdx, "should find cancel item")

	m.Cursor = cancelIdx
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "cancel", m.Selected, "should select cancel")
}

func TestMenu_AutomationIncludesRunLedger(t *testing.T) {
	m := NewMenuModel()

	found := false
	for _, section := range m.Sections {
		if section.Title != "Automation" {
			continue
		}
		for _, category := range section.Categories {
			if category.ID == "runledger" {
				found = true
				assert.Equal(t, "RunLedger", category.Title)
				break
			}
		}
	}

	assert.True(t, found, "Automation section should include RunLedger category")
}

func TestEditor_CtrlC_AlwaysQuits(t *testing.T) {
	tests := []struct {
		give string
		step EditorStep
	}{
		{give: "welcome", step: StepWelcome},
		{give: "menu", step: StepMenu},
		{give: "form", step: StepForm},
		{give: "providers_list", step: StepProvidersList},
		{give: "auth_providers_list", step: StepAuthProvidersList},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			e := NewEditor()
			e.step = tt.step

			model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			ed := model.(*Editor)

			assert.True(t, ed.Cancelled, "ctrl+c should set Cancelled")
			assert.NotNil(t, cmd, "ctrl+c should return quit cmd")
		})
	}
}
