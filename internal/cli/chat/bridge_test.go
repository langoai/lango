package chat

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/turnrunner"
)

// mockSender captures tea.Msg values sent through the msgSender interface.
type mockSender struct {
	msgs []tea.Msg
}

func (m *mockSender) Send(msg tea.Msg) {
	m.msgs = append(m.msgs, msg)
}

func TestEnrichRequest_NilSender(t *testing.T) {
	req := &turnrunner.Request{}
	enrichRequest(nil, req) // must not panic

	assert.Nil(t, req.OnToolCall, "OnToolCall should remain nil when sender is nil")
	assert.Nil(t, req.OnToolResult, "OnToolResult should remain nil when sender is nil")
	assert.Nil(t, req.OnThinking, "OnThinking should remain nil when sender is nil")
	assert.Nil(t, req.OnDelegation, "OnDelegation should remain nil when sender is nil")
	assert.Nil(t, req.OnBudgetWarning, "OnBudgetWarning should remain nil when sender is nil")
}

func TestEnrichRequest_SetsCallbacks(t *testing.T) {
	sender := &mockSender{}
	req := &turnrunner.Request{}

	enrichRequest(sender, req)

	assert.NotNil(t, req.OnToolCall, "OnToolCall should be set")
	assert.NotNil(t, req.OnToolResult, "OnToolResult should be set")
	assert.NotNil(t, req.OnThinking, "OnThinking should be set")
	assert.NotNil(t, req.OnDelegation, "OnDelegation should be set")
	assert.NotNil(t, req.OnBudgetWarning, "OnBudgetWarning should be set")
}

func TestEnrichRequest_PreservesExistingOnChunk(t *testing.T) {
	sender := &mockSender{}
	var called bool
	originalOnChunk := func(chunk string) { called = true }

	req := &turnrunner.Request{
		OnChunk: originalOnChunk,
	}

	enrichRequest(sender, req)

	require.NotNil(t, req.OnChunk, "OnChunk should still be set")
	req.OnChunk("test")
	assert.True(t, called, "original OnChunk should still be callable")
}

func TestEnrichRequest_OnToolCallSendsMsg(t *testing.T) {
	sender := &mockSender{}
	req := &turnrunner.Request{}

	enrichRequest(sender, req)

	params := map[string]any{"path": "/tmp/test.txt"}
	req.OnToolCall("call1", "fs_read", params)

	require.Len(t, sender.msgs, 1, "expected exactly one message")

	msg, ok := sender.msgs[0].(ToolStartedMsg)
	require.True(t, ok, "expected ToolStartedMsg, got %T", sender.msgs[0])
	assert.Equal(t, "call1", msg.CallID)
	assert.Equal(t, "fs_read", msg.ToolName)
	assert.Equal(t, params, msg.Params)
}

func TestEnrichRequest_OnThinkingBoundary(t *testing.T) {
	sender := &mockSender{}
	req := &turnrunner.Request{}

	enrichRequest(sender, req)

	// Start thinking.
	req.OnThinking("agent", true, "thinking...")
	require.Len(t, sender.msgs, 1)

	startMsg, ok := sender.msgs[0].(ThinkingStartedMsg)
	require.True(t, ok, "expected ThinkingStartedMsg, got %T", sender.msgs[0])
	assert.Equal(t, "agent", startMsg.AgentName)
	assert.Equal(t, "thinking...", startMsg.Summary)

	// Small delay so Duration is non-zero.
	time.Sleep(5 * time.Millisecond)

	// End thinking.
	req.OnThinking("agent", false, "summary")
	require.Len(t, sender.msgs, 2)

	finishMsg, ok := sender.msgs[1].(ThinkingFinishedMsg)
	require.True(t, ok, "expected ThinkingFinishedMsg, got %T", sender.msgs[1])
	assert.Equal(t, "agent", finishMsg.AgentName)
	assert.Equal(t, "summary", finishMsg.Summary)
	assert.Greater(t, finishMsg.Duration, time.Duration(0), "Duration should be non-zero")
}

func TestEnrichRequest_OnDelegationSendsMsg(t *testing.T) {
	tests := []struct {
		give       string
		giveFrom   string
		giveTo     string
		giveReason string
		wantFrom   string
		wantTo     string
		wantReason string
	}{
		{
			give:       "delegation with reason",
			giveFrom:   "orchestrator",
			giveTo:     "specialist",
			giveReason: "needs code review",
			wantFrom:   "orchestrator",
			wantTo:     "specialist",
			wantReason: "needs code review",
		},
		{
			give:       "delegation without reason",
			giveFrom:   "planner",
			giveTo:     "coder",
			giveReason: "",
			wantFrom:   "planner",
			wantTo:     "coder",
			wantReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			sender := &mockSender{}
			req := &turnrunner.Request{}

			enrichRequest(sender, req)
			req.OnDelegation(tt.giveFrom, tt.giveTo, tt.giveReason)

			require.Len(t, sender.msgs, 1, "expected exactly one message")

			msg, ok := sender.msgs[0].(DelegationMsg)
			require.True(t, ok, "expected DelegationMsg, got %T", sender.msgs[0])
			assert.Equal(t, tt.wantFrom, msg.From)
			assert.Equal(t, tt.wantTo, msg.To)
			assert.Equal(t, tt.wantReason, msg.Reason)
		})
	}
}

func TestEnrichRequest_OnBudgetWarningSendsMsg(t *testing.T) {
	tests := []struct {
		give    string
		giveUsd int
		giveMax int
		wantUsd int
		wantMax int
	}{
		{
			give:    "80 percent threshold",
			giveUsd: 12,
			giveMax: 15,
			wantUsd: 12,
			wantMax: 15,
		},
		{
			give:    "custom budget max",
			giveUsd: 24,
			giveMax: 30,
			wantUsd: 24,
			wantMax: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			sender := &mockSender{}
			req := &turnrunner.Request{}

			enrichRequest(sender, req)
			req.OnBudgetWarning(tt.giveUsd, tt.giveMax)

			require.Len(t, sender.msgs, 1, "expected exactly one message")

			msg, ok := sender.msgs[0].(BudgetWarningMsg)
			require.True(t, ok, "expected BudgetWarningMsg, got %T", sender.msgs[0])
			assert.Equal(t, tt.wantUsd, msg.Used)
			assert.Equal(t, tt.wantMax, msg.Max)
		})
	}
}
