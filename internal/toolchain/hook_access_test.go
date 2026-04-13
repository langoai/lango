package toolchain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
)

func TestAgentAccessControlHook_Pre(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		allowedTools map[string]map[string]bool
		deniedTools  map[string]map[string]bool
		agentName    string
		toolName     string
		wantAction   PreHookAction
		wantReason   string
	}{
		{
			give:       "no agent name allows all",
			agentName:  "",
			toolName:   "exec",
			wantAction: Continue,
		},
		{
			give:       "unconfigured agent allows all",
			agentName:  "unknown_agent",
			toolName:   "exec",
			wantAction: Continue,
		},
		{
			give: "allowed tool passes through",
			allowedTools: map[string]map[string]bool{
				"researcher": {"web_search": true, "fs_read": true},
			},
			agentName:  "researcher",
			toolName:   "web_search",
			wantAction: Continue,
		},
		{
			give: "disallowed tool is blocked",
			allowedTools: map[string]map[string]bool{
				"researcher": {"web_search": true},
			},
			agentName:  "researcher",
			toolName:   "exec",
			wantAction: Block,
			wantReason: "agent 'researcher' does not have access to tool 'exec'",
		},
		{
			give: "denied tool takes precedence over allowed",
			allowedTools: map[string]map[string]bool{
				"researcher": {"exec": true},
			},
			deniedTools: map[string]map[string]bool{
				"researcher": {"exec": true},
			},
			agentName:  "researcher",
			toolName:   "exec",
			wantAction: Block,
			wantReason: "agent 'researcher' is denied access to tool 'exec'",
		},
		{
			give: "denied tool blocks even without allow list",
			deniedTools: map[string]map[string]bool{
				"planner": {"exec": true},
			},
			agentName:  "planner",
			toolName:   "exec",
			wantAction: Block,
			wantReason: "agent 'planner' is denied access to tool 'exec'",
		},
		{
			give: "empty allow list means no restrictions",
			allowedTools: map[string]map[string]bool{
				"researcher": {},
			},
			agentName:  "researcher",
			toolName:   "exec",
			wantAction: Continue,
		},
		{
			give: "different agent is not affected",
			allowedTools: map[string]map[string]bool{
				"researcher": {"web_search": true},
			},
			agentName:  "executor",
			toolName:   "exec",
			wantAction: Continue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			hook := &AgentAccessControlHook{
				AllowedTools: tt.allowedTools,
				DeniedTools:  tt.deniedTools,
			}

			result, err := hook.Pre(HookContext{
				ToolName:  tt.toolName,
				AgentName: tt.agentName,
				Ctx:       context.Background(),
			})

			require.NoError(t, err)
			assert.Equal(t, tt.wantAction, result.Action)
			if tt.wantReason != "" {
				assert.Equal(t, tt.wantReason, result.BlockReason)
			}
		})
	}
}

func TestAgentAccessControlHook_Metadata(t *testing.T) {
	t.Parallel()

	hook := &AgentAccessControlHook{}
	assert.Equal(t, "agent_access_control", hook.Name())
	assert.Equal(t, 20, hook.Priority())
}

func TestAgentAccessControlHook_DynamicAllowedTools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		dynAllowed   []string
		toolName     string
		agentName    string
		wantAction   PreHookAction
		wantReason   string
	}{
		{
			give:       "allowed tool passes through",
			dynAllowed: []string{"fs_read", "web_search"},
			toolName:   "fs_read",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "disallowed tool is blocked",
			dynAllowed: []string{"fs_read"},
			toolName:   "exec",
			agentName:  "worker",
			wantAction: Block,
			wantReason: "tool restricted by DynamicAllowedTools",
		},
		{
			give:       "runtime essential tool_output_get always allowed",
			dynAllowed: []string{"fs_read"},
			toolName:   "tool_output_get",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "runtime essential builtin_list always allowed",
			dynAllowed: []string{"fs_read"},
			toolName:   "builtin_list",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "runtime essential builtin_search always allowed",
			dynAllowed: []string{"fs_read"},
			toolName:   "builtin_search",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "runtime essential builtin_health always allowed",
			dynAllowed: []string{"fs_read"},
			toolName:   "builtin_health",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "builtin_invoke is NOT a runtime essential — blocked",
			dynAllowed: []string{"fs_read"},
			toolName:   "builtin_invoke",
			agentName:  "worker",
			wantAction: Block,
			wantReason: "tool restricted by DynamicAllowedTools",
		},
		{
			give:       "empty dynAllowed means no restriction",
			dynAllowed: []string{},
			toolName:   "exec",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "nil dynAllowed means no restriction",
			dynAllowed: nil,
			toolName:   "exec",
			agentName:  "worker",
			wantAction: Continue,
		},
		{
			give:       "no agent name skips dynamic check entirely",
			dynAllowed: []string{"fs_read"},
			toolName:   "exec",
			agentName:  "",
			wantAction: Continue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			hook := &AgentAccessControlHook{}

			ctx := context.Background()
			if tt.dynAllowed != nil {
				ctx = ctxkeys.WithDynamicAllowedTools(ctx, tt.dynAllowed)
			}

			result, err := hook.Pre(HookContext{
				ToolName:  tt.toolName,
				AgentName: tt.agentName,
				Ctx:       ctx,
			})

			require.NoError(t, err)
			assert.Equal(t, tt.wantAction, result.Action)
			if tt.wantReason != "" {
				assert.Equal(t, tt.wantReason, result.BlockReason)
			}
		})
	}
}
