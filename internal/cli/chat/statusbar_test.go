package chat

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

func TestRenderHelpBar_AllStates(t *testing.T) {
	tests := []struct {
		give  string
		state chatState
	}{
		{give: "idle", state: stateIdle},
		{give: "streaming", state: stateStreaming},
		{give: "approving", state: stateApproving},
		{give: "cancelling", state: stateCancelling},
		{give: "failed", state: stateFailed},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			output := renderHelpBar(tt.state, 120)
			assert.NotEmpty(t, output)
		})
	}
}

func TestRenderHelpBar_NarrowWidth(t *testing.T) {
	// Regression test for B1: help bar must respect width parameter.
	output := renderHelpBar(stateIdle, 30)
	assert.LessOrEqual(t, lipgloss.Width(output), 30)
}

func TestRenderHelpBar_ZeroWidth(t *testing.T) {
	assert.NotPanics(t, func() {
		renderHelpBar(stateIdle, 0)
	})
}

func TestRenderHelpBar_ContainsCorrectKeys(t *testing.T) {
	idleOutput := renderHelpBar(stateIdle, 120)
	assert.Contains(t, idleOutput, "Enter")
	assert.Contains(t, idleOutput, "x2")
	assert.Contains(t, idleOutput, "Ctrl+D")

	streamingOutput := renderHelpBar(stateStreaming, 120)
	assert.Contains(t, streamingOutput, "Ctrl+C")

	approvingOutput := renderHelpBar(stateApproving, 120)
	assert.Contains(t, approvingOutput, "d/Esc")
	assert.Contains(t, approvingOutput, "Ctrl+D")
}

func TestTurnStateCopy_AllStates(t *testing.T) {
	tests := []struct {
		give  string
		state chatState
	}{
		{give: "idle", state: stateIdle},
		{give: "streaming", state: stateStreaming},
		{give: "approving", state: stateApproving},
		{give: "cancelling", state: stateCancelling},
		{give: "failed", state: stateFailed},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			label, hint, color := turnStateCopy(tt.state)
			assert.NotEmpty(t, label)
			_ = hint  // hint may be empty for some states
			_ = color // color is always set
		})
	}
}

func TestTurnStateCopy_Default(t *testing.T) {
	label, _, _ := turnStateCopy(chatState(99))
	assert.Equal(t, "Ready", label)
}

func TestTurnStateCopy_ApprovingMentionsDenyAndQuit(t *testing.T) {
	label, hint, _ := turnStateCopy(stateApproving)
	assert.Equal(t, "Approval Required", label)
	assert.Contains(t, hint, "d/Esc")
	assert.Contains(t, hint, "Ctrl+D")
}

func TestRenderHeader_NormalWidth(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-4"

	output := renderHeader(cfg, "abc123", 80)
	assert.Contains(t, output, "Lango")
}

func TestRenderHeader_NarrowWidth(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-4"

	assert.NotPanics(t, func() {
		output := renderHeader(cfg, "abc123", 20)
		assert.LessOrEqual(t, lipgloss.Width(output), 20)
		assert.Equal(t, 1, lipgloss.Height(output))
	})
}

func TestRenderHeader_ExtremeNarrowWidth(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-4"

	output := renderHeader(cfg, "abc123", 8)
	assert.LessOrEqual(t, lipgloss.Width(output), 8)
	assert.Equal(t, 1, lipgloss.Height(output))
}

func TestRenderHeader_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}

	output := renderHeader(cfg, "abc123", 80)
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "auto")
}

func TestRenderHeader_NilConfig(t *testing.T) {
	assert.NotPanics(t, func() {
		output := renderHeader(nil, "abc123", 80)
		assert.Contains(t, output, "default")
		assert.Contains(t, output, "auto")
	})
}

func TestRenderHeader_SetupRequiredHidesDefaultReadyProfile(t *testing.T) {
	output := renderHeaderWithSetup(config.DefaultConfig(), "abc123", 80, true)

	assert.Contains(t, output, "Setup Required")
	assert.NotContains(t, output, "default")
	assert.NotContains(t, output, "auto")
}

func TestRenderHeader_SanitizesDisplayFields(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agent.Provider = "\x1b[31mopen\nai\x1b[0m"
	cfg.Agent.Model = "gpt\t4"

	output := renderHeader(cfg, "abc\n123\x1b[0m", 120)
	assert.Contains(t, output, "open ai")
	assert.Contains(t, output, "gpt 4")
	assert.Contains(t, output, "session: abc 123")
	assert.NotContains(t, output, "\x1b[31m")
	assert.Equal(t, 1, lipgloss.Height(output))
}

func TestRenderTurnStrip_AllStates(t *testing.T) {
	tests := []struct {
		give  string
		state chatState
	}{
		{give: "idle", state: stateIdle},
		{give: "streaming", state: stateStreaming},
		{give: "approving", state: stateApproving},
		{give: "cancelling", state: stateCancelling},
		{give: "failed", state: stateFailed},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.NotPanics(t, func() {
				renderTurnStrip(tt.state, 80)
			})
		})
	}
}

func TestRenderTurnStrip_NarrowWidth(t *testing.T) {
	assert.NotPanics(t, func() {
		output := renderTurnStrip(stateIdle, 10)
		assert.LessOrEqual(t, lipgloss.Width(output), 10)
		assert.Equal(t, 1, lipgloss.Height(output))
	})
}

func TestRenderTurnStrip_ContainsLabel(t *testing.T) {
	output := renderTurnStrip(stateIdle, 80)
	assert.Contains(t, output, "Ready")
}

func TestRenderTurnStrip_SetupRequiredReplacesReadyCopy(t *testing.T) {
	output := renderTurnStripWithSetup(stateIdle, 80, true)

	assert.Contains(t, output, "Setup Required")
	assert.Contains(t, output, "lango onboard")
	assert.NotContains(t, output, "Ready")
	assert.NotContains(t, output, "Enter sends")
}

func TestRenderHelpBar_SetupRequiredKeepsCommandsDiscoverable(t *testing.T) {
	output := renderHelpBarWithSetup(stateIdle, 120, true)

	assert.Contains(t, output, "lango onboard")
	assert.Contains(t, output, "lango settings")
	assert.Contains(t, output, "lango doctor")
	assert.Contains(t, output, "/help")
	assert.NotContains(t, output, "Enter")
	assert.LessOrEqual(t, lipgloss.Width(output), 120)
}

func TestRenderTurnStrip_ApprovingNarrowWidth(t *testing.T) {
	output := renderTurnStrip(stateApproving, 14)
	assert.LessOrEqual(t, lipgloss.Width(output), 14)
	assert.Equal(t, 1, lipgloss.Height(output))
}
