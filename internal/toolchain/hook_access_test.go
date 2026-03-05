package toolchain

import (
	"context"
	"testing"
)

func TestAgentAccessControlHook_Pre(t *testing.T) {
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
			hook := &AgentAccessControlHook{
				AllowedTools: tt.allowedTools,
				DeniedTools:  tt.deniedTools,
			}

			result, err := hook.Pre(HookContext{
				ToolName:  tt.toolName,
				AgentName: tt.agentName,
				Ctx:       context.Background(),
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

func TestAgentAccessControlHook_Metadata(t *testing.T) {
	hook := &AgentAccessControlHook{}
	if hook.Name() != "agent_access_control" {
		t.Errorf("Name() = %q, want %q", hook.Name(), "agent_access_control")
	}
	if hook.Priority() != 20 {
		t.Errorf("Priority() = %d, want 20", hook.Priority())
	}
}
