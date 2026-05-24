package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/turnrunner"
)

type failingSubmitExecutor struct {
	err error
}

func (f failingSubmitExecutor) RunStreamingDetailed(
	context.Context,
	string,
	string,
	adk.ChunkCallback,
	...adk.RunOption,
) (adk.RunReport, error) {
	return adk.RunReport{}, f.err
}

func TestUpdateSubmissionFailureRendersErrorAndResetsTurnState(t *testing.T) {
	runnerErr := errors.New("runner unavailable")
	m := New(Deps{
		TurnRunner: turnrunner.New(turnrunner.Config{}, failingSubmitExecutor{err: runnerErr}, submitTestSessionStore{}, nil),
		Config:     readyRemoteConfig(),
		SessionKey: "test-session",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 96, Height: 28})
	m = updated.(*ChatModel)
	m.SetComposerValue("run failing turn")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ChatModel)
	require.NotNil(t, cmd)

	for _, msg := range collectImmediateMsgs(cmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	out := ansi.Strip(m.View())
	assert.Contains(t, out, "run failing turn")
	assert.Contains(t, out, "An error occurred: runner unavailable")
	assert.Contains(t, out, "Last Turn Failed")
}

func TestUpdateChunkDismissesPendingIndicatorAndRendersStream(t *testing.T) {
	m := newTestModel()
	m.pending.Activate()
	require.NotEmpty(t, m.RenderParts().Pending)

	updated, cmd := m.Update(ChunkMsg{Chunk: "streamed answer"})
	m = updated.(*ChatModel)

	assert.NotNil(t, cmd)
	assert.Empty(t, m.RenderParts().Pending)
	assert.Contains(t, ansi.Strip(m.View()), "streamed answer")
}

func TestUpdateRendersToolAndThinkingLifecycleEvents(t *testing.T) {
	m := newTestModel()

	updated, cmd := m.Update(ToolStartedMsg{
		CallID:   "call-1",
		ToolName: "shell",
		Params:   map[string]any{"cmd": "go test ./internal/cli/chat"},
	})
	m = updated.(*ChatModel)
	require.Nil(t, cmd)

	updated, cmd = m.Update(ToolFinishedMsg{
		CallID:   "call-1",
		Success:  true,
		Duration: 250 * time.Millisecond,
		Output:   "ok",
	})
	m = updated.(*ChatModel)
	require.Nil(t, cmd)

	updated, cmd = m.Update(ThinkingStartedMsg{Summary: "Inspect chat update flow"})
	m = updated.(*ChatModel)
	require.Nil(t, cmd)

	updated, cmd = m.Update(ThinkingFinishedMsg{Summary: "Found render path", Duration: time.Second})
	m = updated.(*ChatModel)
	require.Nil(t, cmd)

	out := ansi.Strip(m.RenderParts().Main)
	assert.Contains(t, out, "shell")
	assert.Contains(t, out, "go test ./internal/cli/chat")
	assert.Contains(t, out, "ok")
	assert.Contains(t, out, "Thinking")
}

func TestUpdateRendersProgressAndCrossChannelEvents(t *testing.T) {
	m := newTestModel()

	messages := []tea.Msg{
		WarningMsg{Elapsed: 3 * time.Second, HardCeiling: 10 * time.Second},
		DelegationMsg{From: "planner", To: "coder", Reason: "Split implementation work"},
		BudgetWarningMsg{Used: 4, Max: 5},
		RecoveryMsg{Action: "retry", CauseClass: "transient_network", Attempt: 2, Backoff: time.Second},
		ChannelMessageMsg{Channel: "telegram", SenderName: "Ada", Text: "remote request", SessionKey: "remote-session"},
		SystemMsg{Text: "system notice"},
	}
	for _, msg := range messages {
		updated, cmd := m.Update(msg)
		m = updated.(*ChatModel)
		require.Nil(t, cmd)
	}

	out := ansi.Strip(m.RenderParts().Main)
	assert.Contains(t, out, "Approaching timeout")
	assert.Contains(t, out, "3s / 10s")
	assert.Contains(t, out, "planner")
	assert.Contains(t, out, "coder")
	assert.Contains(t, out, "Split implementation work")
	assert.Contains(t, out, "Delegation budget: 4/5")
	assert.Contains(t, out, "transient_network")
	assert.Contains(t, out, "remote request")
	assert.Contains(t, out, "system notice")
}

func TestUpdateAccumulatesTokenUsageAndFlushesTurnSummaryOnDone(t *testing.T) {
	var gotSummary TurnTokenUsageMsg
	m := newTestModel()
	m.onTurnSummary = func(_ string, msg TurnTokenUsageMsg) {
		gotSummary = msg
	}
	m.SetComposerValue("count this turn")

	cmd := m.SubmitComposerWithParent(context.Background())
	require.NotNil(t, cmd)

	updated, updateCmd := m.Update(TokenUsageTeaMsg{
		InputTokens:      7,
		OutputTokens:     11,
		CacheTokens:      3,
		EstimatedCostUSD: 0.04,
	})
	m = updated.(*ChatModel)
	require.Nil(t, updateCmd)

	updated, doneCmd := m.Update(DoneMsg{Result: turnrunner.Result{Outcome: "success", ResponseText: "finished"}})
	m = updated.(*ChatModel)
	for _, msg := range collectImmediateMsgs(doneCmd) {
		updated, _ = m.Update(msg)
		m = updated.(*ChatModel)
	}

	assert.Equal(t, int64(7), gotSummary.InputTokens)
	assert.Equal(t, int64(11), gotSummary.OutputTokens)
	assert.Equal(t, int64(18), gotSummary.TotalTokens)
	assert.Equal(t, int64(3), gotSummary.CacheTokens)
	assert.Equal(t, 0.04, gotSummary.EstimatedCostUSD)
	assert.Contains(t, ansi.Strip(m.RenderParts().Main), "18 total")
}

func TestViewIncludesPendingAndApprovalSections(t *testing.T) {
	shared := &stubSharedPendingStore{
		latest: &ApprovalRequestMsg{
			Request: approval.ApprovalRequest{
				ID:       "apr-render",
				ToolName: "fs_write",
				Summary:  "Review write",
			},
			Response: make(chan approval.ApprovalResponse, 1),
		},
	}
	m := newTestModelWithSharedPending(shared)
	m.state = stateApproving
	m.pending.Activate()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(*ChatModel)

	view := ansi.Strip(m.View())
	assert.Contains(t, view, "Working")
	assert.Contains(t, view, "Tool Approval Required")
	assert.Contains(t, view, "Review write")
	assert.False(t, strings.Contains(view, "\n\n\n"))
}
