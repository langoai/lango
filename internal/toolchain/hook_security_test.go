package toolchain

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityFilterHook_Pre(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give            string
		blockedPatterns []string
		blockedTools    []string
		toolName        string
		params          map[string]interface{}
		wantAction      PreHookAction
		wantReason      string
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
			t.Parallel()

			hook := &SecurityFilterHook{
				BlockedPatterns: tt.blockedPatterns,
				BlockedTools:    tt.blockedTools,
			}

			result, err := hook.Pre(HookContext{
				ToolName: tt.toolName,
				Params:   tt.params,
				Ctx:      context.Background(),
			})

			require.NoError(t, err)
			assert.Equal(t, tt.wantAction, result.Action)
			if tt.wantReason != "" {
				assert.Equal(t, tt.wantReason, result.BlockReason)
			}
		})
	}
}

func TestSecurityFilterHook_Metadata(t *testing.T) {
	t.Parallel()

	hook := &SecurityFilterHook{}
	assert.Equal(t, "security_filter", hook.Name())
	assert.Equal(t, 10, hook.Priority())
}

func TestDefaultBlockedPatterns(t *testing.T) {
	t.Parallel()

	patterns := DefaultBlockedPatterns()
	assert.True(t, len(patterns) > 0, "default patterns should not be empty")

	// Verify critical patterns are present.
	patternsStr := strings.Join(patterns, "|")
	for _, must := range []string{"rm -rf /", "mkfs.", "dd if=/dev/zero"} {
		assert.Contains(t, patternsStr, must, "default patterns should contain %q", must)
	}
}

func TestNewSecurityFilterHook_MergesDefaults(t *testing.T) {
	t.Parallel()

	hook := NewSecurityFilterHook([]string{"custom_pattern", "rm -rf /"})

	// Should contain defaults + custom pattern.
	assert.Contains(t, hook.BlockedPatterns, "custom_pattern", "user pattern should be included")
	assert.Contains(t, hook.BlockedPatterns, "rm -rf /", "default pattern should be included")

	// Should not duplicate "rm -rf /" (appears in both defaults and user patterns).
	count := 0
	for _, p := range hook.BlockedPatterns {
		if p == "rm -rf /" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate patterns should be deduplicated")
}

func TestSecurityFilterHook_DefaultPatternsBlock(t *testing.T) {
	t.Parallel()

	hook := NewSecurityFilterHook(nil)

	tests := []struct {
		give       string
		wantAction PreHookAction
	}{
		{"rm -rf /", Block},
		{"mkfs.ext4 /dev/sda1", Block},
		{"dd if=/dev/zero of=/dev/sda", Block},
		{"echo hello", Continue},
		{"go build ./...", Continue},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			result, err := hook.Pre(HookContext{
				ToolName: "exec",
				Params:   map[string]interface{}{"command": tt.give},
				Ctx:      context.Background(),
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantAction, result.Action, "command: %q", tt.give)
		})
	}
}
