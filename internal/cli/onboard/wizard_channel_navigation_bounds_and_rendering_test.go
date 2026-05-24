package onboard

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

func TestWizardChannelNavigationBoundsAndRendering(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepChannel)

	model, cmd := w.handleChannelStep(tea.KeyMsg{Type: tea.KeyUp})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, 0, w.channelCursor)

	for range channelOptions {
		model, cmd = w.handleChannelStep(tea.KeyMsg{Type: tea.KeyDown})
		w = model.(*Wizard)
		require.Nil(t, cmd)
	}
	assert.Equal(t, len(channelOptions)-1, w.channelCursor)

	model, cmd = w.handleChannelStep(tea.KeyMsg{Type: tea.KeyDown})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, len(channelOptions)-1, w.channelCursor)

	view := w.View()
	for _, want := range []string{
		"Setup Wizard",
		"Select a Channel",
		"Telegram",
		"Discord",
		"Slack",
		"Configure later in settings",
		"ctrl+n: next",
	} {
		assert.True(t, strings.Contains(view, want), "view should contain %q:\n%s", want, view)
	}
}

func TestWizardSavesChannelAndSecurityConfigUpdates(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepChannel)
	w.channelCursor = 0 // Telegram.

	model, cmd := w.handleChannelStep(tea.KeyMsg{Type: tea.KeyEnter})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	require.Equal(t, StepChannel, w.step)
	require.False(t, w.channelSelectMode)
	require.NotNil(t, w.activeForm)

	setWizardChannelNavigationBoundsAndRenderingWizardFieldValue(t, w, "telegram_token", "123456:token")
	model, cmd = w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	require.Equal(t, StepSecurity, w.step)
	assert.True(t, w.Config().Channels.Telegram.Enabled)
	assert.Equal(t, "123456:token", w.Config().Channels.Telegram.BotToken)

	setWizardBoolField(t, w, "interceptor_enabled", true)
	setWizardBoolField(t, w, "interceptor_pii", true)
	setWizardChannelNavigationBoundsAndRenderingWizardFieldValue(t, w, "interceptor_policy", "all")
	model, cmd = w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	require.Equal(t, StepTest, w.step)
	assert.True(t, w.Config().Security.Interceptor.Enabled)
	assert.True(t, w.Config().Security.Interceptor.RedactPII)
	assert.Equal(t, config.ApprovalPolicy("all"), w.Config().Security.Interceptor.ApprovalPolicy)
	assert.NotEmpty(t, w.testResults)
}

func TestWizardForwardsNonKeyMessagesToActiveForm(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepAgent)

	model, cmd := w.Update(tuicore.FieldOptionsLoadedMsg{
		FieldKey: "model",
		Options:  []string{"gpt-4o", "gpt-4.1"},
	})
	w = model.(*Wizard)
	require.Nil(t, cmd)

	field := wizardFieldByKey(t, w, "model")
	assert.Equal(t, tuicore.InputSearchSelect, field.Type)
	assert.Equal(t, []string{"gpt-4o", "gpt-4.1"}, field.Options)
	assert.Contains(t, field.Description, "Fetched 2 models")

	model, cmd = w.Update(tuicore.FieldOptionsLoadedMsg{
		FieldKey: "model",
		Err:      errors.New("provider unavailable"),
	})
	w = model.(*Wizard)
	require.Nil(t, cmd)

	field = wizardFieldByKey(t, w, "model")
	assert.Equal(t, tuicore.InputText, field.Type)
	assert.Contains(t, field.Description, "provider unavailable")
}

func TestWizardNavigationAndCompletionBranches(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())

	model, cmd := w.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	w = model.(*Wizard)
	require.Nil(t, cmd)
	assert.Equal(t, StepProvider, w.step)
	assert.False(t, w.Completed)
	assert.False(t, w.Cancelled)

	model, cmd = w.handleFormStep(tea.KeyMsg{Type: tea.KeyEsc})
	w = model.(*Wizard)
	assert.True(t, w.Cancelled)
	assert.NotNil(t, cmd)

	w = NewWizard(config.DefaultConfig())
	w.enterStep(StepTest)
	model, cmd = w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	w = model.(*Wizard)
	assert.True(t, w.Completed)
	assert.NotNil(t, cmd)
}

func TestWizardAgentValidationOutcomes(t *testing.T) {
	t.Parallel()

	w := NewWizard(config.DefaultConfig())
	w.enterStep(StepAgent)

	maxTokens := wizardFieldByKey(t, w, "maxtokens")
	require.NotNil(t, maxTokens.Validate)
	require.NoError(t, maxTokens.Validate("4096"))
	require.Error(t, maxTokens.Validate("0"))
	require.Error(t, maxTokens.Validate("many"))

	temp := wizardFieldByKey(t, w, "temp")
	require.NotNil(t, temp.Validate)
	require.NoError(t, temp.Validate("0.7"))
	require.Error(t, temp.Validate("-0.1"))
	require.Error(t, temp.Validate("2.1"))
	require.Error(t, temp.Validate("warm"))
}

func setWizardBoolField(t *testing.T, w *Wizard, key string, checked bool) {
	t.Helper()

	field := wizardFieldByKey(t, w, key)
	field.Checked = checked
	field.Value = ""
	if checked {
		field.Value = "true"
	}
}

func setWizardChannelNavigationBoundsAndRenderingWizardFieldValue(t *testing.T, w *Wizard, key, value string) {
	t.Helper()

	field := wizardFieldByKey(t, w, key)
	field.Value = value
	if field.TextInput.Value() != value {
		field.TextInput.SetValue(value)
	}
}

func wizardFieldByKey(t *testing.T, w *Wizard, key string) *tuicore.Field {
	t.Helper()
	require.NotNil(t, w.activeForm)

	for _, field := range w.activeForm.Fields {
		if field.Key == key {
			return field
		}
	}
	t.Fatalf("missing wizard field %q", key)
	return nil
}
