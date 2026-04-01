package exec

import (
	"strings"
	"unicode"

	"mvdan.cc/sh/v3/syntax"
)

// detectOpaquePattern checks if the command contains patterns that prevent
// static analysis of what will actually execute. Returns the first matching
// ReasonCode, or ReasonNone if the command is transparent.
//
// safeVars is the set of environment variable names that are considered safe
// (e.g., HOME, PATH). Variable references to these do not trigger detection.
func detectOpaquePattern(cmd string, safeVars map[string]struct{}) ReasonCode {
	// 1. Command substitution: $(...) or backticks.
	if strings.Contains(cmd, "$(") || strings.ContainsRune(cmd, '`') {
		return ReasonCmdSubstitution
	}

	// 2. Unsafe variable expansion: ${VAR} or $VAR where VAR is not in safe set.
	if reason := checkUnsafeVarExpansion(cmd, safeVars); reason != ReasonNone {
		return reason
	}

	// 3. Eval verb.
	if extractVerb(cmd) == "eval" {
		return ReasonEvalVerb
	}

	// 4. Encoded pipe: base64 decode piped to shell/eval.
	if detectEncodedPipe(cmd) {
		return ReasonEncodedPipe
	}

	return ReasonNone
}

// checkUnsafeVarExpansion scans for ${VAR} and $VAR patterns where the
// variable name is not in the safe set.
func checkUnsafeVarExpansion(cmd string, safeVars map[string]struct{}) ReasonCode {
	for i := 0; i < len(cmd); i++ {
		if cmd[i] != '$' {
			continue
		}
		if i+1 >= len(cmd) {
			continue
		}

		next := cmd[i+1]

		// Skip $( — handled by command substitution check.
		if next == '(' {
			continue
		}

		var varName string
		if next == '{' {
			// ${VAR} form.
			end := strings.IndexByte(cmd[i+2:], '}')
			if end < 0 {
				continue
			}
			varName = cmd[i+2 : i+2+end]
			// Strip parameter expansion operators: ${VAR:-default}, ${VAR:+alt}, etc.
			for _, sep := range []string{":-", ":+", ":?", ":", "-", "+", "?"} {
				if idx := strings.Index(varName, sep); idx >= 0 {
					varName = varName[:idx]
				}
			}
		} else if isVarStart(next) {
			// $VAR form: collect contiguous alphanumeric/underscore chars.
			end := i + 2
			for end < len(cmd) && isVarChar(cmd[end]) {
				end++
			}
			varName = cmd[i+1 : end]
		} else {
			// $? $$ $! $# etc. — special variables, not a risk.
			continue
		}

		if varName == "" {
			continue
		}

		if _, safe := safeVars[varName]; !safe {
			return ReasonUnsafeVarExpand
		}
	}
	return ReasonNone
}

func isVarStart(b byte) bool {
	return b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isVarChar(b byte) bool {
	return isVarStart(b) || (b >= '0' && b <= '9')
}

// detectEncodedPipe checks for patterns like "base64 -d | bash" or
// "base64 --decode | sh" that could hide malicious payloads.
// Handles multi-pipe chains (e.g., "cat file | base64 -d | eval").
func detectEncodedPipe(cmd string) bool {
	lower := strings.ToLower(cmd)

	// Look for base64 decode somewhere in the command.
	if !strings.Contains(lower, "base64") {
		return false
	}
	if !strings.Contains(lower, "-d") && !strings.Contains(lower, "--decode") {
		return false
	}

	// Check all pipe segments for a shell/eval target.
	parts := strings.Split(lower, "|")
	for i := 1; i < len(parts); i++ {
		segment := strings.TrimSpace(parts[i])
		firstWord := segment
		if spaceIdx := strings.IndexFunc(segment, unicode.IsSpace); spaceIdx >= 0 {
			firstWord = segment[:spaceIdx]
		}
		switch firstWord {
		case "sh", "bash", "zsh", "dash", "eval":
			return true
		}
	}
	return false
}

// detectShellConstruct checks if the command contains shell constructs that
// prevent full static analysis. Uses AST-based detection with string-based
// fallback. Returns the matching ReasonCode, or ReasonNone if transparent.
//
// Detected constructs:
//   - Heredoc (<<, <<<): arbitrary code possible, content not analyzable
//   - Process substitution (<(cmd), >(cmd)): opaque subshell execution
//   - Grouped subshell ((cmd; cmd), { cmd; }): multiple commands, partial analysis
//   - Shell function definition (f() { ... }): function body can contain anything
func detectShellConstruct(cmd string) ReasonCode {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return ReasonNone
	}

	// Try AST-based detection first.
	parser := syntax.NewParser()
	f, err := parser.Parse(strings.NewReader(trimmed), "")
	if err == nil {
		if reason := detectShellConstructAST(f); reason != ReasonNone {
			return reason
		}
		return ReasonNone
	}

	// Fallback: string-based detection for when AST parsing fails.
	return detectShellConstructString(trimmed)
}

// detectShellConstructAST walks the parsed AST to find opaque shell constructs.
func detectShellConstructAST(f *syntax.File) ReasonCode {
	var reason ReasonCode
	syntax.Walk(f, func(node syntax.Node) bool {
		if reason != ReasonNone {
			return false // already found a match, stop walking
		}
		switch n := node.(type) {
		case *syntax.Redirect:
			// Heredoc: << or <<< operators.
			if n.Op == syntax.Hdoc || n.Op == syntax.DashHdoc || n.Op == syntax.WordHdoc {
				reason = ReasonHeredoc
				return false
			}
		case *syntax.ProcSubst:
			// Process substitution: <(cmd) or >(cmd).
			reason = ReasonProcessSubst
			return false
		case *syntax.Subshell:
			// Grouped subshell: (cmd; cmd).
			reason = ReasonGroupedSubshell
			return false
		case *syntax.Block:
			// Brace group: { cmd; }.
			// Only flag if the file contains more than a simple command —
			// standalone brace groups are suspicious.
			reason = ReasonGroupedSubshell
			return false
		case *syntax.FuncDecl:
			// Shell function definition: f() { ... }.
			reason = ReasonShellFunction
			return false
		}
		return true
	})
	return reason
}

// detectShellConstructString is the string-based fallback for detecting shell
// constructs when AST parsing fails.
func detectShellConstructString(cmd string) ReasonCode {
	// Heredoc: << pattern (not <<< which is here-string, also opaque).
	if strings.Contains(cmd, "<<") {
		return ReasonHeredoc
	}

	// Process substitution: <( or >( patterns.
	if strings.Contains(cmd, "<(") || strings.Contains(cmd, ">(") {
		return ReasonProcessSubst
	}

	// Shell function definition: pattern like "name()" or "name ()" followed by {.
	if strings.Contains(cmd, "() {") || strings.Contains(cmd, "(){") {
		return ReasonShellFunction
	}

	return ReasonNone
}
