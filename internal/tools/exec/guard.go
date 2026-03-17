package exec

import (
	"os"
	"path/filepath"
	"strings"
)

// CommandGuard checks shell commands for access to protected paths and
// dangerous process management operations.
type CommandGuard struct {
	protectedPaths []string // absolute paths to protect
	homeDir        string
	replacer       *strings.Replacer // pre-built for normalizeCommand
}

// NewCommandGuard creates a CommandGuard that blocks commands targeting the given
// protected paths. Paths are resolved to absolute form at construction time.
func NewCommandGuard(protectedPaths []string) *CommandGuard {
	home, _ := os.UserHomeDir()

	resolved := make([]string, 0, len(protectedPaths))
	for _, p := range protectedPaths {
		abs := expandAndAbs(p, home)
		if abs != "" {
			resolved = append(resolved, abs)
		}
	}

	// Pre-build replacer for normalizeCommand (single-pass replacement).
	// Includes $HOME, ${HOME}, and tilde-at-word-boundary variants.
	var replacer *strings.Replacer
	if home != "" {
		replacer = strings.NewReplacer(
			"${HOME}", home,
			"$HOME", home,
			" ~/", " "+home+"/",
			"\t~/", "\t"+home+"/",
			"\"~/", "\""+home+"/",
			"'~/", "'"+home+"/",
		)
	}

	return &CommandGuard{
		protectedPaths: resolved,
		homeDir:        home,
		replacer:       replacer,
	}
}

// processKillVerbs are command verbs that manage/terminate processes.
var processKillVerbs = map[string]struct{}{
	"kill":    {},
	"pkill":   {},
	"killall": {},
}

// CheckCommand returns true and a reason string if the command should be blocked.
// It checks for:
//  1. Commands that access protected paths (sqlite3, cat, python, etc.)
//  2. Process management commands (kill, pkill, killall)
func (g *CommandGuard) CheckCommand(command string) (blocked bool, reason string) {
	normalized := g.normalizeCommand(command)

	// Check process kill verbs.
	verb := extractVerb(normalized)
	if _, isKill := processKillVerbs[verb]; isKill {
		return true, "process management commands (" + verb + ") are blocked for security — " +
			"use exec_stop to stop background processes started by exec_bg"
	}

	// Check protected path access.
	for _, protPath := range g.protectedPaths {
		if strings.Contains(normalized, protPath) {
			return true, "command accesses protected data path (" + protPath + ") — " +
				"use the built-in tools (settings, secrets_get, etc.) instead of direct file/DB access"
		}
	}

	return false, ""
}

// normalizeCommand expands ~ and $HOME references to the actual home directory
// so that path matching works regardless of how the user references the path.
func (g *CommandGuard) normalizeCommand(cmd string) string {
	if g.homeDir == "" {
		return cmd
	}

	// Single-pass replacement for $HOME, ${HOME}, and tilde-at-word-boundary.
	result := g.replacer.Replace(cmd)

	// Handle tilde at start of string (not covered by Replacer which needs a prefix char).
	if strings.HasPrefix(result, "~/") {
		result = g.homeDir + result[1:]
	}

	return result
}

// extractVerb returns the first command word (the executable name).
// Strips any path prefix (e.g., /usr/bin/kill → kill).
func extractVerb(cmd string) string {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return ""
	}
	// Find end of first word.
	end := strings.IndexByte(trimmed, ' ')
	if end < 0 {
		end = len(trimmed)
	}
	verb := trimmed[:end]
	verb = filepath.Base(verb)
	return strings.ToLower(verb)
}

// expandAndAbs expands ~ and converts to absolute path.
func expandAndAbs(path, homeDir string) string {
	if homeDir != "" && strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}
	path = strings.ReplaceAll(path, "$HOME", homeDir)
	path = strings.ReplaceAll(path, "${HOME}", homeDir)

	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return filepath.Clean(abs)
}
