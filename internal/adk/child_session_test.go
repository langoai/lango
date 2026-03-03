package adk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

func TestStructuredSummarizer(t *testing.T) {
	tests := []struct {
		name     string
		give     []session.Message
		wantText string
	}{
		{
			name: "last assistant message",
			give: []session.Message{
				{Role: types.RoleUser, Content: "do something"},
				{Role: types.RoleAssistant, Content: "first response"},
				{Role: types.RoleUser, Content: "more"},
				{Role: types.RoleAssistant, Content: "final response"},
			},
			wantText: "final response",
		},
		{
			name: "no assistant messages",
			give: []session.Message{
				{Role: types.RoleUser, Content: "hello"},
			},
			wantText: "",
		},
		{
			name:     "empty messages",
			give:     nil,
			wantText: "",
		},
	}

	s := &StructuredSummarizer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Summarize(tt.give)
			require.NoError(t, err)
			assert.Equal(t, tt.wantText, got)
		})
	}
}

func TestChildSessionContext(t *testing.T) {
	ctx := context.Background()

	_, ok := ChildSessionFromContext(ctx)
	assert.False(t, ok, "empty context should return false")

	info := ChildSessionInfo{
		ChildKey:  "child-1",
		ParentKey: "parent-1",
		AgentName: "operator",
	}
	ctx = WithChildSession(ctx, info)

	got, ok := ChildSessionFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, info, got)
}
