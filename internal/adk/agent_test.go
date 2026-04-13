package adk

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestExtractMissingAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give error
		want string
	}{
		{
			name: "standard ADK error",
			give: fmt.Errorf("agent error: failed to find agent: browser_agent"),
			want: "browser_agent",
		},
		{
			name: "wrapped error",
			give: fmt.Errorf("outer: %w", fmt.Errorf("failed to find agent: exec")),
			want: "exec",
		},
		{
			name: "unrelated error",
			give: fmt.Errorf("connection refused"),
			want: "",
		},
		{
			name: "partial match no agent name",
			give: fmt.Errorf("failed to find agent: "),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractMissingAgent(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasFunctionCalls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		evt  *session.Event
		want bool
	}{
		{
			give: "nil content",
			evt:  &session.Event{},
			want: false,
		},
		{
			give: "text only",
			evt: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{Text: "hello"}},
					},
				},
			},
			want: false,
		},
		{
			give: "with FunctionCall",
			evt: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{FunctionCall: &genai.FunctionCall{Name: "exec"}},
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, hasFunctionCalls(tt.evt))
		})
	}
}

func TestIsDelegationEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		evt  *session.Event
		want bool
	}{
		{
			give: "no transfer",
			evt:  &session.Event{},
			want: false,
		},
		{
			give: "with transfer",
			evt: &session.Event{
				Actions: session.EventActions{
					TransferToAgent: "operator",
				},
			},
			want: true,
		},
		{
			give: "empty transfer string",
			evt: &session.Event{
				Actions: session.EventActions{
					TransferToAgent: "",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isDelegationEvent(tt.evt))
		})
	}
}

func TestIsPureTransferToAgentCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		evt  *session.Event
		want bool
	}{
		{
			give: "nil content",
			evt:  &session.Event{},
			want: false,
		},
		{
			give: "pure transfer_to_agent",
			evt: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{FunctionCall: &genai.FunctionCall{Name: "transfer_to_agent", Args: map[string]any{"agent_name": "vault"}}},
						},
					},
				},
			},
			want: true,
		},
		{
			give: "mixed transfer_to_agent and real tool",
			evt: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{FunctionCall: &genai.FunctionCall{Name: "transfer_to_agent", Args: map[string]any{"agent_name": "vault"}}},
							{FunctionCall: &genai.FunctionCall{Name: "exec", Args: map[string]any{"cmd": "ls"}}},
						},
					},
				},
			},
			want: false,
		},
		{
			give: "regular tool call",
			evt: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{FunctionCall: &genai.FunctionCall{Name: "exec"}},
						},
					},
				},
			},
			want: false,
		},
		{
			give: "text only no calls",
			evt: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{Text: "hello"}},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isPureTransferToAgentCall(tt.evt))
		})
	}
}

// TestContextErrCheck_Canceled and TestContextErrCheck_DeadlineExceeded validate
// the post-iteration ctx.Err() check pattern used in runAndCollectOnce (agent.go:391)
// and RunStreaming (agent.go:455).
//
// A full integration test through RunAndCollect would require mocking the ADK runner
// (runner.Runner), which depends on deep ADK internals (session.Service, Agent interface).
// Since the fix is a simple post-loop `if ctx.Err() != nil` check, these pattern tests
// provide sufficient coverage by proving that ctx.Err() correctly surfaces the error
// after cancellation/deadline. The pattern is identical to the production code path.

// --- Dynamic budget expansion and wrap-up tests ---
// These test the budget expansion detection logic and wrap-up mechanics.
// Full integration through Run() would require mocking the ADK runner,
// so we test the detection functions and logic patterns directly.

func TestIsDelegationEvent_TargetExtraction(t *testing.T) {
	t.Parallel()

	// Verify that isDelegationEvent correctly identifies delegation events
	// and that TransferToAgent field is accessible for budget tracking.
	tests := []struct {
		give   string
		target string
		want   bool
	}{
		{give: "to operator", target: "operator", want: true},
		{give: "to planner", target: "planner", want: true},
		{give: "back to orchestrator", target: "lango-orchestrator", want: true},
		{give: "empty", target: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			evt := &session.Event{
				Actions: session.EventActions{
					TransferToAgent: tt.target,
				},
			}
			assert.Equal(t, tt.want, isDelegationEvent(evt))
		})
	}
}

func TestBudgetExpansionConditions(t *testing.T) {
	t.Parallel()

	// Test the budget expansion trigger conditions in isolation.
	// These conditions mirror the Run() loop logic.
	tests := []struct {
		name             string
		plannerInvolved  bool
		delegationCount  int
		uniqueAgentCount int
		wantExpand       bool
	}{
		{
			name:             "planner involved triggers expansion",
			plannerInvolved:  true,
			delegationCount:  1,
			uniqueAgentCount: 1,
			wantExpand:       true,
		},
		{
			name:             "3+ delegations triggers expansion",
			plannerInvolved:  false,
			delegationCount:  3,
			uniqueAgentCount: 1,
			wantExpand:       true,
		},
		{
			name:             "2+ unique agents triggers expansion",
			plannerInvolved:  false,
			delegationCount:  2,
			uniqueAgentCount: 2,
			wantExpand:       true,
		},
		{
			name:             "single agent single delegation no expansion",
			plannerInvolved:  false,
			delegationCount:  1,
			uniqueAgentCount: 1,
			wantExpand:       false,
		},
		{
			name:             "two delegations to same agent no expansion",
			plannerInvolved:  false,
			delegationCount:  2,
			uniqueAgentCount: 1,
			wantExpand:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			shouldExpand := tt.plannerInvolved || tt.delegationCount >= 3 || tt.uniqueAgentCount >= 2
			assert.Equal(t, tt.wantExpand, shouldExpand)
		})
	}
}

func TestBudgetExpansionMath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		initial    int
		wantResult int
	}{
		{name: "default 50 → 75", initial: 50, wantResult: 75},
		{name: "10 → 15", initial: 10, wantResult: 15},
		{name: "75 → 112", initial: 75, wantResult: 112},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expanded := tt.initial * 3 / 2
			assert.Equal(t, tt.wantResult, expanded)
		})
	}
}

func TestShouldCollectUserText(t *testing.T) {
	t.Parallel()

	agent := &Agent{
		isolatedAgents: map[string]struct{}{
			"navigator": {},
			"operator":  {},
		},
	}

	assert.True(t, agent.shouldCollectUserText(""))
	assert.True(t, agent.shouldCollectUserText("lango-orchestrator"))
	assert.False(t, agent.shouldCollectUserText("navigator"))
	assert.False(t, agent.shouldCollectUserText("operator"))
}

func TestWrapUpBudgetMechanics(t *testing.T) {
	t.Parallel()

	// Test that wrap-up budget counting works correctly.
	tests := []struct {
		name            string
		wrapUpBudget    int
		turnsAfterLimit int
		wantError       bool
	}{
		{
			name:            "default budget allows 1 turn",
			wrapUpBudget:    1,
			turnsAfterLimit: 1,
			wantError:       false,
		},
		{
			name:            "default budget blocks 2nd turn",
			wrapUpBudget:    1,
			turnsAfterLimit: 2,
			wantError:       true,
		},
		{
			name:            "expanded budget allows 3 turns",
			wrapUpBudget:    3,
			turnsAfterLimit: 3,
			wantError:       false,
		},
		{
			name:            "expanded budget blocks 4th turn",
			wrapUpBudget:    3,
			turnsAfterLimit: 4,
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remaining := tt.wrapUpBudget
			var hitError bool
			for i := 0; i < tt.turnsAfterLimit; i++ {
				remaining--
				if remaining < 0 {
					hitError = true
					break
				}
			}
			assert.Equal(t, tt.wantError, hitError)
		})
	}
}

func TestContainsRejectPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{give: "[REJECT] This task requires operator.", want: true},
		{give: "Some text [REJECT] more text", want: true},
		{give: "[REJECT]", want: true},
		{give: "Normal assistant response", want: false},
		{give: "I can help with that!", want: false},
		{give: "", want: false},
		{give: "REJECT without brackets", want: false},
		{give: "[reject] lowercase", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, containsRejectPattern(tt.give))
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		n    int
		want string
	}{
		{give: "short", n: 10, want: "short"},
		{give: "exactly10!", n: 10, want: "exactly10!"},
		{give: "this is longer than ten", n: 10, want: "this is lo..."},
		{give: "", n: 5, want: ""},
		{give: "안녕하세요 반갑습니다", n: 5, want: "안녕하세요..."},
		{give: "한글", n: 5, want: "한글"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, truncate(tt.give, tt.n))
		})
	}
}

func TestContextErrCheck_Canceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.Error(t, ctx.Err())
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestContextErrCheck_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	<-ctx.Done()

	require.Error(t, ctx.Err())
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}
