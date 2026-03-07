package toolchain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKnowledgeSaver implements KnowledgeSaver for testing.
type mockKnowledgeSaver struct {
	calls []knowledgeSaveCall
	err   error
}

type knowledgeSaveCall struct {
	sessionKey string
	toolName   string
	params     map[string]interface{}
	result     interface{}
}

func (m *mockKnowledgeSaver) SaveToolResult(_ context.Context, sessionKey, toolName string, params map[string]interface{}, result interface{}) error {
	m.calls = append(m.calls, knowledgeSaveCall{
		sessionKey: sessionKey,
		toolName:   toolName,
		params:     params,
		result:     result,
	})
	return m.err
}

func TestKnowledgeSaveHook_Post(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give          string
		saveableTools []string
		toolName      string
		toolErr       error
		wantSaved     bool
	}{
		{
			give:          "saves result for saveable tool",
			saveableTools: []string{"web_search", "fs_read"},
			toolName:      "web_search",
			wantSaved:     true,
		},
		{
			give:          "skips non-saveable tool",
			saveableTools: []string{"web_search"},
			toolName:      "exec",
			wantSaved:     false,
		},
		{
			give:          "skips failed tool execution",
			saveableTools: []string{"web_search"},
			toolName:      "web_search",
			toolErr:       errors.New("search failed"),
			wantSaved:     false,
		},
		{
			give:          "empty saveable list saves nothing",
			saveableTools: nil,
			toolName:      "exec",
			wantSaved:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			saver := &mockKnowledgeSaver{}
			hook := NewKnowledgeSaveHook(saver, tt.saveableTools)

			err := hook.Post(HookContext{
				ToolName:   tt.toolName,
				SessionKey: "session-1",
				Params:     map[string]interface{}{"q": "test"},
				Ctx:        context.Background(),
			}, "search-result", tt.toolErr)

			require.NoError(t, err)

			saved := len(saver.calls) > 0
			assert.Equal(t, tt.wantSaved, saved)

			if tt.wantSaved && len(saver.calls) == 1 {
				call := saver.calls[0]
				assert.Equal(t, tt.toolName, call.toolName)
				assert.Equal(t, "session-1", call.sessionKey)
				assert.Equal(t, "search-result", call.result)
			}
		})
	}
}

func TestKnowledgeSaveHook_Post_SaverError(t *testing.T) {
	t.Parallel()

	saverErr := errors.New("db write failed")
	saver := &mockKnowledgeSaver{err: saverErr}
	hook := NewKnowledgeSaveHook(saver, []string{"web_search"})

	err := hook.Post(HookContext{
		ToolName: "web_search",
		Ctx:      context.Background(),
	}, "result", nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, saverErr)
}

func TestKnowledgeSaveHook_Metadata(t *testing.T) {
	t.Parallel()

	hook := NewKnowledgeSaveHook(&mockKnowledgeSaver{}, nil)
	assert.Equal(t, "knowledge_save", hook.Name())
	assert.Equal(t, 100, hook.Priority())
}
