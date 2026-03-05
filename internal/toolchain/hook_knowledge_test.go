package toolchain

import (
	"context"
	"errors"
	"testing"
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
			saver := &mockKnowledgeSaver{}
			hook := NewKnowledgeSaveHook(saver, tt.saveableTools)

			err := hook.Post(HookContext{
				ToolName:   tt.toolName,
				SessionKey: "session-1",
				Params:     map[string]interface{}{"q": "test"},
				Ctx:        context.Background(),
			}, "search-result", tt.toolErr)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			saved := len(saver.calls) > 0
			if saved != tt.wantSaved {
				t.Errorf("saved = %v, want %v", saved, tt.wantSaved)
			}

			if tt.wantSaved && len(saver.calls) == 1 {
				call := saver.calls[0]
				if call.toolName != tt.toolName {
					t.Errorf("toolName = %q, want %q", call.toolName, tt.toolName)
				}
				if call.sessionKey != "session-1" {
					t.Errorf("sessionKey = %q, want %q", call.sessionKey, "session-1")
				}
				if call.result != "search-result" {
					t.Errorf("result = %v, want %q", call.result, "search-result")
				}
			}
		})
	}
}

func TestKnowledgeSaveHook_Post_SaverError(t *testing.T) {
	saverErr := errors.New("db write failed")
	saver := &mockKnowledgeSaver{err: saverErr}
	hook := NewKnowledgeSaveHook(saver, []string{"web_search"})

	err := hook.Post(HookContext{
		ToolName: "web_search",
		Ctx:      context.Background(),
	}, "result", nil)

	if err == nil {
		t.Fatal("expected error from saver failure")
	}
	if !errors.Is(err, saverErr) {
		t.Errorf("err = %v, want wrapping %v", err, saverErr)
	}
}

func TestKnowledgeSaveHook_Metadata(t *testing.T) {
	hook := NewKnowledgeSaveHook(&mockKnowledgeSaver{}, nil)
	if hook.Name() != "knowledge_save" {
		t.Errorf("Name() = %q, want %q", hook.Name(), "knowledge_save")
	}
	if hook.Priority() != 100 {
		t.Errorf("Priority() = %d, want 100", hook.Priority())
	}
}
