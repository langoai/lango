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

	streamingOutput := renderHelpBar(stateStreaming, 120)
	assert.Contains(t, streamingOutput, "Ctrl+C")
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
		_ = output
	})
}

func TestRenderHeader_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}

	output := renderHeader(cfg, "abc123", 80)
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "auto")
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
		_ = output
	})
}

func TestRenderTurnStrip_ContainsLabel(t *testing.T) {
	output := renderTurnStrip(stateIdle, 80)
	assert.Contains(t, output, "Ready")
}
