package toolchain

import "strings"

// SecurityFilterHook blocks dangerous command patterns before tool execution.
// Priority: 10 (runs early to reject bad requests fast).
type SecurityFilterHook struct {
	// BlockedPatterns contains substrings that cause a tool invocation to be blocked.
	// Matched case-insensitively against the "command" parameter of exec-like tools.
	BlockedPatterns []string

	// BlockedTools contains tool names that are unconditionally blocked.
	BlockedTools []string
}

// NewSecurityFilterHook creates a SecurityFilterHook with the given blocked command patterns.
func NewSecurityFilterHook(blockedPatterns []string) *SecurityFilterHook {
	return &SecurityFilterHook{BlockedPatterns: blockedPatterns}
}

// Compile-time interface check.
var _ PreToolHook = (*SecurityFilterHook)(nil)

// Name returns the hook name.
func (h *SecurityFilterHook) Name() string { return "security_filter" }

// Priority returns 10 (high priority — runs early).
func (h *SecurityFilterHook) Priority() int { return 10 }

// Pre checks whether the tool invocation should be blocked based on
// tool name blocklist and dangerous command patterns.
func (h *SecurityFilterHook) Pre(ctx HookContext) (PreHookResult, error) {
	// Check unconditionally blocked tools.
	for _, blocked := range h.BlockedTools {
		if ctx.ToolName == blocked {
			return PreHookResult{
				Action:      Block,
				BlockReason: "tool '" + ctx.ToolName + "' is blocked by security policy",
			}, nil
		}
	}

	// Check command patterns for exec-like tools.
	cmd, ok := ctx.Params["command"].(string)
	if !ok || cmd == "" {
		return PreHookResult{Action: Continue}, nil
	}

	cmdLower := strings.ToLower(cmd)
	for _, pattern := range h.BlockedPatterns {
		if strings.Contains(cmdLower, strings.ToLower(pattern)) {
			return PreHookResult{
				Action:      Block,
				BlockReason: "command matches blocked pattern: " + pattern,
			}, nil
		}
	}

	return PreHookResult{Action: Continue}, nil
}
