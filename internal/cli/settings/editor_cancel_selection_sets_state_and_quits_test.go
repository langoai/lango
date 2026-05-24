package settings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestEditorCancelSelectionSetsStateAndQuits(t *testing.T) {
	e := NewEditorWithConfig(config.DefaultConfig())

	cmd := e.handleMenuSelection("cancel")

	assert.True(t, e.Cancelled)
	require.Error(t, e.err)
	assert.EqualError(t, e.err, "settings cancelled")
	require.NotNil(t, cmd)
}

func TestEditorSetupFlowEscCancelsAndReturnsToMenu(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.handleMenuSelection("smartaccount")
	require.Equal(t, StepForm, e.step)

	e.startSetupFlow()
	require.Equal(t, StepSetupFlow, e.step)
	require.NotNil(t, e.setupFlow)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed := model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepMenu, ed.step)
	assert.Nil(t, ed.setupFlow)
}

func TestEditorWindowSizeUpdatesRenderDimensions(t *testing.T) {
	e := NewEditor()

	model, cmd := e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	ed := model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, 120, ed.width)
	assert.Equal(t, 40, ed.height)
	assert.Contains(t, ed.View(), "Settings")
}
