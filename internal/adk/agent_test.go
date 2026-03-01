package adk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestExtractMissingAgent(t *testing.T) {
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
			got := extractMissingAgent(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasFunctionCalls(t *testing.T) {
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
			assert.Equal(t, tt.want, hasFunctionCalls(tt.evt))
		})
	}
}

func TestIsDelegationEvent(t *testing.T) {
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
			assert.Equal(t, tt.want, isDelegationEvent(tt.evt))
		})
	}
}
