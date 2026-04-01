package exec

import (
	"path/filepath"
	"strings"
)

// shellWrappers are command verbs that invoke a shell with -c flag.
var shellWrappers = map[string]struct{}{
	"sh":   {},
	"bash": {},
	"zsh":  {},
	"dash": {},
}

// isShellWrapper returns true if the verb (after path stripping and lowercasing)
// is a known shell wrapper binary.
func isShellWrapper(verb string) bool {
	_, ok := shellWrappers[verb]
	return ok
}

// unwrapShellWrapper detects and unwraps one level of sh -c / bash -c wrapper.
// Returns the inner command and true if unwrapped, or the original command and false.
//
// Supported patterns:
//   - sh -c "cmd", bash -c 'cmd', /bin/sh -c cmd, /usr/bin/bash -c "cmd"
//   - zsh -c, dash -c
//
// Not supported (returns original):
//   - sh -lc, /usr/bin/env sh -c, nested wrappers, escaped quotes inside
func unwrapShellWrapper(cmd string) (inner string, unwrapped bool) {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return cmd, false
	}

	// Split into whitespace-separated fields for robust parsing.
	fields := strings.Fields(trimmed)
	if len(fields) < 3 {
		// Need at least: <shell> -c <cmd>
		return cmd, false
	}

	// Check if the first field (verb) is a shell wrapper.
	verb := strings.ToLower(filepath.Base(fields[0]))
	if !isShellWrapper(verb) {
		return cmd, false
	}

	// Second field must be exactly "-c".
	if fields[1] != "-c" {
		return cmd, false
	}

	// Extract the inner command: everything after "<shell> -c ".
	// Find the position of "-c" after the first word to avoid matching
	// "-c" inside the shell binary path (e.g., /opt/bash-cfg/sh).
	shellEnd := strings.IndexFunc(trimmed, func(r rune) bool { return r == ' ' || r == '\t' })
	if shellEnd < 0 {
		return cmd, false
	}
	flagIdx := strings.Index(trimmed[shellEnd:], "-c")
	if flagIdx < 0 {
		return cmd, false
	}
	flagIdx += shellEnd // adjust to absolute position
	afterFlag := trimmed[flagIdx+2:]
	inner = strings.TrimSpace(afterFlag)
	if inner == "" {
		return cmd, false
	}

	inner = stripQuotes(inner)
	if inner == "" {
		return cmd, false
	}

	return inner, true
}

// stripQuotes removes matching outer single or double quotes from a string.
// Only strips if the first and last characters are the same quote character.
func stripQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	first := s[0]
	last := s[len(s)-1]
	if (first == '"' || first == '\'') && first == last {
		return s[1 : len(s)-1]
	}
	return s
}
