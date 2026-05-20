package adk

import (
	"context"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestPureTransferCallClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event *session.Event
		want  bool
	}{
		{
			name:  "nil content",
			event: &session.Event{},
			want:  false,
		},
		{
			name: "text only",
			event: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{Parts: []*genai.Part{{Text: "route"}}},
				},
			},
			want: false,
		},
		{
			name: "only transfer calls",
			event: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{Name: "transfer_to_agent"}},
						{FunctionCall: &genai.FunctionCall{Name: "transfer_to_agent"}},
					}},
				},
			},
			want: true,
		},
		{
			name: "mixed transfer and real tool call",
			event: &session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{Name: "transfer_to_agent"}},
						{FunctionCall: &genai.FunctionCall{Name: "read_file"}},
					}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isPureTransferToAgentCall(tt.event))
		})
	}
}

func TestRunAllowsPureTransferButRejectsRootToolCalls(t *testing.T) {
	t.Parallel()

	child := newStaticADKAgent(t, "operator", nil)

	transferRoot := newPureTransferCallClassificationSingleEventAgent(t, "transfer-root", []adkagent.Agent{child}, &session.Event{
		Author: "transfer-root",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{
					Name: "transfer_to_agent",
					Args: map[string]any{"agent_name": "operator"},
				}},
			}},
		},
	})
	transferAgent, err := NewAgentFromADK(transferRoot, newMockStore())
	require.NoError(t, err)

	var transferEvents int
	for event, err := range transferAgent.Run(context.Background(), "pureTransferCallClassification9-transfer", "delegate") {
		require.NoError(t, err)
		require.NotNil(t, event)
		transferEvents++
	}
	assert.Equal(t, 1, transferEvents)

	toolRoot := newPureTransferCallClassificationSingleEventAgent(t, "tool-root", []adkagent.Agent{child}, &session.Event{
		Author: "tool-root",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{Name: "read_file"}},
			}},
		},
	})
	toolAgent, err := NewAgentFromADK(toolRoot, newMockStore())
	require.NoError(t, err)

	for event, err := range toolAgent.Run(context.Background(), "pureTransferCallClassification9-root-tool", "call tool") {
		require.Error(t, err)
		assert.Nil(t, event)
		assert.ErrorContains(t, err, `orchestrator emitted direct tool call "read_file"`)
		return
	}
	t.Fatal("expected direct root tool call to stop the run")
}

func TestMissingAgentRetryHelpers(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "planner", extractMissingAgent(pureTransferCallClassificationErrorString("failed to find agent: planner")))
	assert.Empty(t, extractMissingAgent(pureTransferCallClassificationErrorString("provider failed")))

	assert.True(t, shouldRetryMissingAgent("planner", 0))
	assert.True(t, shouldRetryMissingAgent("operator", 2))
	assert.False(t, shouldRetryMissingAgent("missing-worker", 0))
	assert.False(t, shouldRetryMissingAgent("", 3))

	builtin := buildMissingAgentCorrection("planner", []string{"operator"}, containsBuiltinTargetName)
	assert.Contains(t, builtin, `Built-in agent "planner" does not exist`)
	assert.NotContains(t, builtin, "Valid agents")

	external := buildMissingAgentCorrection("missing-worker", []string{"operator", "vault"}, nil)
	assert.Contains(t, external, `Agent "missing-worker" does not exist`)
	assert.Contains(t, external, "Valid agents: operator, vault")
}

func TestToolSignatureFallsBackWhenArgsCannotMarshal(t *testing.T) {
	t.Parallel()

	event := &session.Event{
		Author: "operator",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{
					Name: "bad_args",
					Args: map[string]any{"unsupported": make(chan int)},
				}},
			}},
		},
	}

	assert.Equal(t, "operator|bad_args|{}", extractPrimaryToolSignature(event))
}

func newPureTransferCallClassificationSingleEventAgent(
	t *testing.T,
	name string,
	subAgents []adkagent.Agent,
	event *session.Event,
) adkagent.Agent {
	t.Helper()

	agent, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: name + " pureTransferCallClassification9 test agent",
		SubAgents:   subAgents,
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				yield(event, nil)
			}
		},
	})
	require.NoError(t, err)
	return agent
}

type pureTransferCallClassificationErrorString string

func (e pureTransferCallClassificationErrorString) Error() string { return string(e) }
