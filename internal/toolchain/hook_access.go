package toolchain

import "github.com/langoai/lango/internal/ctxkeys"

// runtimeEssentials are system tools always allowed regardless of DynamicAllowedTools.
// builtin_invoke is deliberately excluded: it can proxy-execute other tools, bypassing the allowlist.
var runtimeEssentials = map[string]bool{
	"tool_output_get": true,
	"builtin_list":    true,
	"builtin_search":  true,
	"builtin_health":  true,
}

// AgentAccessControlHook enforces per-agent tool ACL.
// Priority: 20 (runs after security filter but before execution).
type AgentAccessControlHook struct {
	// AllowedTools maps agent name → set of allowed tool names.
	// An empty or missing entry means the agent has no restrictions (all tools allowed).
	AllowedTools map[string]map[string]bool

	// DeniedTools maps agent name → set of denied tool names.
	// Deny takes precedence over allow.
	DeniedTools map[string]map[string]bool
}

// NewAgentAccessControlHook creates an AgentAccessControlHook.
// Pass nil for allowedTools to start with no restrictions.
func NewAgentAccessControlHook(allowedTools map[string]map[string]bool) *AgentAccessControlHook {
	return &AgentAccessControlHook{AllowedTools: allowedTools}
}

// Compile-time interface check.
var _ PreToolHook = (*AgentAccessControlHook)(nil)

// Name returns the hook name.
func (h *AgentAccessControlHook) Name() string { return "agent_access_control" }

// Priority returns 20.
func (h *AgentAccessControlHook) Priority() int { return 20 }

// Pre checks whether the current agent is allowed to use the tool.
func (h *AgentAccessControlHook) Pre(ctx HookContext) (PreHookResult, error) {
	agentName := ctx.AgentName
	if agentName == "" {
		// No agent context — allow (backwards compatible with non-agent execution).
		return PreHookResult{Action: Continue}, nil
	}

	// Check deny list first (takes precedence).
	if denied, ok := h.DeniedTools[agentName]; ok {
		if denied[ctx.ToolName] {
			return PreHookResult{
				Action:      Block,
				BlockReason: "agent '" + agentName + "' is denied access to tool '" + ctx.ToolName + "'",
			}, nil
		}
	}

	// Check allow list — if configured, agent can only use listed tools.
	if allowed, ok := h.AllowedTools[agentName]; ok && len(allowed) > 0 {
		if !allowed[ctx.ToolName] {
			return PreHookResult{
				Action:      Block,
				BlockReason: "agent '" + agentName + "' does not have access to tool '" + ctx.ToolName + "'",
			}, nil
		}
	}

	// Check context-level dynamic tool restrictions.
	if dynAllowed := ctxkeys.DynamicAllowedToolsFromContext(ctx.Ctx); len(dynAllowed) > 0 {
		if runtimeEssentials[ctx.ToolName] {
			return PreHookResult{Action: Continue}, nil
		}
		allowSet := make(map[string]bool, len(dynAllowed))
		for _, t := range dynAllowed {
			allowSet[t] = true
		}
		if !allowSet[ctx.ToolName] {
			return PreHookResult{
				Action:      Block,
				BlockReason: "tool restricted by DynamicAllowedTools",
			}, nil
		}
	}

	return PreHookResult{Action: Continue}, nil
}
