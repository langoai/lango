package onboard

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestWizardUpdateHandlesWindowSizeAndCancel(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())

	model, cmd := w.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	updated := model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, 100, updated.width)
	assert.Equal(t, 40, updated.height)

	model, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	updated = model.(*Wizard)
	assert.True(t, updated.Cancelled)
	assert.NotNil(t, cmd)
}

func TestWizardNavigationSavesProviderAndAgentForms(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	w := NewWizard(cfg)

	setWizardFieldValue(t, w, "id", "openai")
	model, cmd := w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	require.Equal(t, StepAgent, w.step)
	assert.Equal(t, "openai", w.Config().Agent.Provider)

	setWizardFieldValue(t, w, "provider", "anthropic")
	setWizardFieldValue(t, w, "model", "claude-sonnet-4-5-20250929")
	setWizardFieldValue(t, w, "maxtokens", "4096")
	setWizardFieldValue(t, w, "temp", "0.4")
	model, cmd = w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	require.Equal(t, StepChannel, w.step)
	assert.Equal(t, "anthropic", w.Config().Agent.Provider)
	assert.Equal(t, "claude-sonnet-4-5-20250929", w.Config().Agent.Model)
	assert.Equal(t, 4096, w.Config().Agent.MaxTokens)
	assert.InDelta(t, 0.4, w.Config().Agent.Temperature, 0.0001)
}

func TestWizardChannelSelectionEnablesChannelAndEscapeReturnsToSelector(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepChannel)
	w.channelCursor = 1 // Discord.

	model, cmd := w.handleChannelStep(tea.KeyMsg{Type: tea.KeyEnter})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, string(channelOptions[1].ID), w.channelChoice)
	assert.True(t, w.Config().Channels.Discord.Enabled)
	assert.False(t, w.channelSelectMode)
	require.NotNil(t, w.activeForm)

	model, cmd = w.handleChannelStep(tea.KeyMsg{Type: tea.KeyEsc})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.True(t, w.channelSelectMode)
	assert.Nil(t, w.activeForm)

	model, cmd = w.handleChannelStep(tea.KeyMsg{Type: tea.KeyEsc})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, StepAgent, w.step)
}

func TestWizardSkipChannelAdvancesWithoutEnablingChannels(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepChannel)
	w.channelCursor = len(channelOptions) - 1

	model, cmd := w.handleChannelStep(tea.KeyMsg{Type: tea.KeyEnter})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, StepSecurity, w.step)
	assert.Equal(t, "skip", w.channelChoice)
	assert.False(t, w.Config().Channels.Telegram.Enabled)
	assert.False(t, w.Config().Channels.Discord.Enabled)
	assert.False(t, w.Config().Channels.Slack.Enabled)
}

func TestWizardTestStepViewAndCompletion(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepTest)
	w.testResults = []TestResult{
		{Name: "Provider configured", Status: "pass", Message: "ok"},
		{Name: "API key set", Status: "warn", Message: "placeholder"},
		{Name: "Agent model set", Status: "fail", Message: "missing"},
	}

	view := w.View()
	for _, want := range []string{"Provider configured", "API key set", "Agent model set"} {
		assert.True(t, strings.Contains(view, want), "view should contain %q:\n%s", want, view)
	}

	model, cmd := w.handleTestStep(tea.KeyMsg{Type: tea.KeyEnter})
	w = model.(*Wizard)
	assert.True(t, w.Completed)
	assert.NotNil(t, cmd)
}

func setWizardFieldValue(t *testing.T, w *Wizard, key, value string) {
	t.Helper()
	require.NotNil(t, w.activeForm)

	for _, field := range w.activeForm.Fields {
		if field.Key != key {
			continue
		}
		field.Value = value
		if field.TextInput.Value() != value {
			field.TextInput.SetValue(value)
		}
		return
	}
	t.Fatalf("missing wizard field %q", key)
}
