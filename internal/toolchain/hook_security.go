package toolchain

import "strings"

// SecurityFilterHook blocks dangerous command patterns before tool execution.
// Priority: 10 (runs early to reject bad requests fast).
type SecurityFilterHook struct {
	// BlockedPatterns contains the original-case patterns (for error messages).
	BlockedPatterns []string

	// blockedPatternsLower contains pre-lowercased patterns for matching.
	blockedPatternsLower []string

	// BlockedTools contains tool names that are unconditionally blocked.
	BlockedTools []string
}

// DefaultBlockedPatterns returns dangerous command patterns that are always blocked
// regardless of user configuration. These represent catastrophic operations that
// should never be executed by an AI agent.
func DefaultBlockedPatterns() []string {
	return []string{
		"rm -rf /",
		"mkfs.",
		"dd if=/dev/zero",
		":(){ :|:& };:",
		"> /dev/sda",
		"chmod -R 777 /",
		"dd if=/dev/random",
		"mv / ",
		"> /dev/null 2>&1 &",
	}
}

// NewSecurityFilterHook creates a SecurityFilterHook with default dangerous patterns
// merged with the given user-configured blocked patterns.
func NewSecurityFilterHook(blockedPatterns []string) *SecurityFilterHook {
	merged := DefaultBlockedPatterns()
	// Append user patterns, deduplicating against defaults.
	seen := make(map[string]bool, len(merged))
	for _, p := range merged {
		seen[strings.ToLower(p)] = true
	}
	for _, p := range blockedPatterns {
		if !seen[strings.ToLower(p)] {
			merged = append(merged, p)
		}
	}
	// Pre-lowercase all patterns for fast matching in Pre().
	lower := make([]string, len(merged))
	for i, p := range merged {
		lower[i] = strings.ToLower(p)
	}
	return &SecurityFilterHook{BlockedPatterns: merged, blockedPatternsLower: lower}
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

	// Use pre-lowercased patterns when available (via constructor).
	if len(h.blockedPatternsLower) == len(h.BlockedPatterns) && len(h.blockedPatternsLower) > 0 {
		for i, patternLower := range h.blockedPatternsLower {
			if strings.Contains(cmdLower, patternLower) {
				return PreHookResult{
					Action:      Block,
					BlockReason: "command matches blocked pattern: " + h.BlockedPatterns[i],
				}, nil
			}
		}
	} else {
		for _, pattern := range h.BlockedPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				return PreHookResult{
					Action:      Block,
					BlockReason: "command matches blocked pattern: " + pattern,
				}, nil
			}
		}
	}

	return PreHookResult{Action: Continue}, nil
}
