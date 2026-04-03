package toolchain

import "strings"

// SecurityFilterHook blocks dangerous command patterns before tool execution.
// Priority: 10 (runs early to reject bad requests fast).
type SecurityFilterHook struct {
	// BlockedPatterns contains the original-case patterns (for error messages).
	BlockedPatterns []string

	// blockedPatternsLower contains pre-lowercased patterns for matching.
	blockedPatternsLower []string

	// ObservePatterns contains patterns that trigger observe-level logging.
	// These commands are allowed but flagged as common obfuscation vectors.
	ObservePatterns []string

	// observePatternsLower contains pre-lowercased observe patterns for matching.
	observePatternsLower []string

	// compoundPatterns contains multi-part patterns pre-computed at construction.
	compoundPatterns []compoundPattern

	// BlockedTools contains tool names that are unconditionally blocked.
	BlockedTools []string
}

// DefaultBlockedPatterns returns dangerous command patterns that are always blocked
// regardless of user configuration. These represent catastrophic operations that
// should never be executed by an AI agent.
func DefaultBlockedPatterns() []string {
	return []string{
		// Original patterns — filesystem destruction and device writes.
		"rm -rf /",
		"mkfs.",
		"dd if=/dev/zero",
		":(){ :|:& };:",
		"> /dev/sda",
		"chmod -r 777 /",
		"dd if=/dev/random",
		"mv / ",
		"> /dev/null 2>&1 &",

		// Privilege escalation.
		"sudo ",
		"su -",
		"chmod +s",
		"chown root",

		// Reverse shell tools.
		"nc -l",
		"ncat ",
		"socat ",

		// Block device writes.
		"dd of=/dev/",
		"tee /dev/sda",

		// Mass deletion / data destruction.
		"shred /",

		// Crontab removal.
		"crontab -r",
	}
}

// compoundPattern requires ALL parts to be present in the command for a match.
// This handles cases where a single substring is insufficient (e.g., "curl" alone
// is fine, "| sh" alone is fine, but together they indicate piped-to-shell RCE).
type compoundPattern struct {
	parts   []string // all parts must be present (lowercased)
	display string   // human-readable description for error messages
}

// defaultCompoundPatterns returns compound patterns that block piped-download
// remote code execution. Each entry requires ALL parts to match.
func defaultCompoundPatterns() []compoundPattern {
	return []compoundPattern{
		{parts: []string{"curl", "| sh"}, display: "curl ... | sh"},
		{parts: []string{"curl", "| bash"}, display: "curl ... | bash"},
		{parts: []string{"wget", "| sh"}, display: "wget ... | sh"},
		{parts: []string{"wget", "| bash"}, display: "wget ... | bash"},
	}
}

// DefaultObservePatterns returns patterns that are legitimate but common
// obfuscation vectors. Commands matching these are allowed to execute but are
// flagged with an Observe action so callers can log or audit them.
func DefaultObservePatterns() []string {
	return []string{
		"python -c",
		"python3 -c",
		"perl -e",
		"node -e",
		"ruby -e",
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

	// Observe patterns — defaults only (no user extension for now).
	observe := DefaultObservePatterns()
	observeLower := make([]string, len(observe))
	for i, p := range observe {
		observeLower[i] = strings.ToLower(p)
	}

	return &SecurityFilterHook{
		BlockedPatterns:      merged,
		blockedPatternsLower: lower,
		ObservePatterns:      observe,
		observePatternsLower: observeLower,
		compoundPatterns:     defaultCompoundPatterns(),
	}
}

// Compile-time interface check.
var _ PreToolHook = (*SecurityFilterHook)(nil)

// Name returns the hook name.
func (h *SecurityFilterHook) Name() string { return "security_filter" }

// Priority returns 10 (high priority — runs early).
func (h *SecurityFilterHook) Priority() int { return 10 }

// matchPattern checks cmdLower against a paired set of original/lowercased patterns.
// Returns the index of the first match, or -1 if none match.
func matchPattern(cmdLower string, originals, lowered []string) int {
	if len(lowered) == len(originals) && len(lowered) > 0 {
		for i, patternLower := range lowered {
			if strings.Contains(cmdLower, patternLower) {
				return i
			}
		}
	} else {
		for i, pattern := range originals {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				return i
			}
		}
	}
	return -1
}

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

	// Check simple blocked patterns.
	if idx := matchPattern(cmdLower, h.BlockedPatterns, h.blockedPatternsLower); idx >= 0 {
		return PreHookResult{
			Action:      Block,
			BlockReason: "command matches blocked pattern: " + h.BlockedPatterns[idx],
		}, nil
	}

	// Check compound patterns (e.g., curl + | sh together).
	for _, cp := range h.compoundPatterns {
		allMatch := true
		for _, part := range cp.parts {
			if !strings.Contains(cmdLower, part) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return PreHookResult{
				Action:      Block,
				BlockReason: "command matches blocked pattern: " + cp.display,
			}, nil
		}
	}

	// Check observe-level patterns (allowed but flagged).
	if idx := matchPattern(cmdLower, h.ObservePatterns, h.observePatternsLower); idx >= 0 {
		return PreHookResult{
			Action:        Observe,
			ObserveReason: "command matches observe pattern: " + h.ObservePatterns[idx],
		}, nil
	}

	return PreHookResult{Action: Continue}, nil
}
