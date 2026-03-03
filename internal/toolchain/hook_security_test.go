package toolchain

import (
	"context"
	"testing"
)

func TestSecurityFilterHook_Pre(t *testing.T) {
	tests := []struct {
		give           string
		blockedPatterns []string
		blockedTools   []string
		toolName       string
		params         map[string]interface{}
		wantAction     PreHookAction
		wantReason     string
	}{
		{
			give:       "allowed tool passes through",
			toolName:   "exec",
			params:     map[string]interface{}{"command": "ls -la"},
			wantAction: Continue,
		},
		{
			give:         "blocked tool is rejected",
			blockedTools: []string{"dangerous_tool"},
			toolName:     "dangerous_tool",
			params:       map[string]interface{}{},
			wantAction:   Block,
			wantReason:   "tool 'dangerous_tool' is blocked by security policy",
		},
		{
			give:            "blocked command pattern is rejected",
			blockedPatterns: []string{"rm -rf", "DROP TABLE"},
			toolName:        "exec",
			params:          map[string]interface{}{"command": "rm -rf /"},
			wantAction:      Block,
			wantReason:      "command matches blocked pattern: rm -rf",
		},
		{
			give:            "pattern matching is case insensitive",
			blockedPatterns: []string{"DROP TABLE"},
			toolName:        "exec",
			params:          map[string]interface{}{"command": "drop table users"},
			wantAction:      Block,
			wantReason:      "command matches blocked pattern: DROP TABLE",
		},
		{
			give:            "safe command passes pattern check",
			blockedPatterns: []string{"rm -rf"},
			toolName:        "exec",
			params:          map[string]interface{}{"command": "echo hello"},
			wantAction:      Continue,
		},
		{
			give:       "no command parameter passes through",
			toolName:   "exec",
			params:     map[string]interface{}{},
			wantAction: Continue,
		},
		{
			give:            "non-exec tool ignores command patterns",
			blockedPatterns: []string{"rm -rf"},
			toolName:        "fs_read",
			params:          map[string]interface{}{"path": "/tmp"},
			wantAction:      Continue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			hook := &SecurityFilterHook{
				BlockedPatterns: tt.blockedPatterns,
				BlockedTools:    tt.blockedTools,
			}

			result, err := hook.Pre(HookContext{
				ToolName: tt.toolName,
				Params:   tt.params,
				Ctx:      context.Background(),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Action != tt.wantAction {
				t.Errorf("Action = %d, want %d", result.Action, tt.wantAction)
			}
			if tt.wantReason != "" && result.BlockReason != tt.wantReason {
				t.Errorf("BlockReason = %q, want %q", result.BlockReason, tt.wantReason)
			}
		})
	}
}

func TestSecurityFilterHook_Metadata(t *testing.T) {
	hook := &SecurityFilterHook{}
	if hook.Name() != "security_filter" {
		t.Errorf("Name() = %q, want %q", hook.Name(), "security_filter")
	}
	if hook.Priority() != 10 {
		t.Errorf("Priority() = %d, want 10", hook.Priority())
	}
}
