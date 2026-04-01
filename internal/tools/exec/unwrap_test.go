package exec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"mvdan.cc/sh/v3/syntax"
)

func TestUnwrapShellWrapper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give          string
		wantInner     string
		wantUnwrapped bool
	}{
		// Supported: sh -c with double quotes
		{give: `sh -c "kill 1234"`, wantInner: "kill 1234", wantUnwrapped: true},
		// Supported: bash -c with single quotes
		{give: `bash -c 'lango security check'`, wantInner: "lango security check", wantUnwrapped: true},
		// Supported: path-prefixed sh
		{give: `/bin/sh -c "cat ~/.lango/keyfile"`, wantInner: "cat ~/.lango/keyfile", wantUnwrapped: true},
		// Supported: long path prefix
		{give: `/usr/bin/bash -c "echo hello"`, wantInner: "echo hello", wantUnwrapped: true},
		// Unquoted: first token only (POSIX semantics)
		{give: `sh -c echo hello`, wantInner: "echo", wantUnwrapped: true},
		// Supported: zsh
		{give: `zsh -c "ls -la"`, wantInner: "ls -la", wantUnwrapped: true},
		// Supported: dash
		{give: `dash -c 'go build ./...'`, wantInner: "go build ./...", wantUnwrapped: true},
		// Not a shell wrapper: python3 -c
		{give: `python3 -c "print('hi')"`, wantInner: `python3 -c "print('hi')"`, wantUnwrapped: false},
		// Not a shell wrapper: sh without -c
		{give: `sh script.sh`, wantInner: `sh script.sh`, wantUnwrapped: false},
		// Supported (P2): login shell flag -lc
		{give: `sh -lc "cmd"`, wantInner: "cmd", wantUnwrapped: true},
		// Supported (P2): env wrapper
		{give: `/usr/bin/env sh -c "cmd"`, wantInner: "cmd", wantUnwrapped: true},
		// Supported (P2): env prefix
		{give: `env bash -c "cmd"`, wantInner: "cmd", wantUnwrapped: true},
		// Edge: empty string
		{give: "", wantInner: "", wantUnwrapped: false},
		// Edge: sh -c with no argument
		{give: "sh -c", wantInner: "sh -c", wantUnwrapped: false},
		// Edge: sh -c with only whitespace after
		{give: "sh -c   ", wantInner: "sh -c   ", wantUnwrapped: false},
		// Not a shell: ls command
		{give: "ls -la", wantInner: "ls -la", wantUnwrapped: false},
		// Supported: tab between sh and -c
		{give: "sh\t-c \"kill 1\"", wantInner: "kill 1", wantUnwrapped: true},
		// P1 fix: positional args after quoted command are ignored
		{give: `bash -c "kill 1234" ignored`, wantInner: "kill 1234", wantUnwrapped: true},
		{give: `sh -c 'lango cron' myname`, wantInner: "lango cron", wantUnwrapped: true},
		{give: `bash -c "echo hello" arg1 arg2`, wantInner: "echo hello", wantUnwrapped: true},
		// P1 fix: unquoted extracts first token only
		{give: `bash -c echo foo bar`, wantInner: "echo", wantUnwrapped: true},
		// P1 fix: unmatched quote → allow-without-unwrap
		{give: `sh -c "kill 1234`, wantInner: `sh -c "kill 1234`, wantUnwrapped: false},
		// P2: login shell with -lc flag and dangerous command
		{give: `sh -lc "kill 1234"`, wantInner: "kill 1234", wantUnwrapped: true},
		// P2: interactive shell with -ic flag
		{give: `bash -ic "echo hello"`, wantInner: "echo hello", wantUnwrapped: true},
		// P2: env wrapper with full path
		{give: `/usr/bin/env sh -c "echo hello"`, wantInner: "echo hello", wantUnwrapped: true},
		// P2: env wrapper with bare env
		{give: `env zsh -c "ls -la"`, wantInner: "ls -la", wantUnwrapped: true},
		// P2: env wrapper with login flag
		{give: `env sh -lc "kill 9999"`, wantInner: "kill 9999", wantUnwrapped: true},
		// P2: nested shell wrapper (recursive unwrap)
		{give: `sh -c "bash -c \"inner cmd\""`, wantInner: "inner cmd", wantUnwrapped: true},
		// P2: double nested
		{give: `sh -c "bash -c \"zsh -c \\\"deep\\\"\""`, wantInner: "deep", wantUnwrapped: true},
		// P2: single quotes inside nested wrapper
		{give: `bash -c 'sh -c "hello world"'`, wantInner: "hello world", wantUnwrapped: true},
		// Fix B: env with variable assignment before shell verb
		{give: `env FOO=1 sh -c "kill 1234"`, wantInner: "kill 1234", wantUnwrapped: true},
		// Fix B: env -i flag before shell verb
		{give: `env -i bash -c "lango cron"`, wantInner: "lango cron", wantUnwrapped: true},
		// Fix B: env -u flag with argument before shell verb
		{give: `env -u SECRET sh -c "echo hi"`, wantInner: "echo hi", wantUnwrapped: true},
		// Fix B: env -C flag with argument before shell verb
		{give: `env -C /tmp bash -c "ls"`, wantInner: "ls", wantUnwrapped: true},
		// Fix B: env -S flag (split-string) with argument before shell verb
		{give: `env -S "FOO=1 BAR=2" sh -c "echo test"`, wantInner: "echo test", wantUnwrapped: true},
		// Fix B: env -- terminator before shell verb
		{give: `env -- sh -c "kill 1"`, wantInner: "kill 1", wantUnwrapped: true},
		// Fix B: env with path-like assignment is not a valid env assignment
		{give: `env ./foo=bar`, wantInner: `env ./foo=bar`, wantUnwrapped: false},
		// Fix B: env -u VAR — VAR must be skipped, not treated as command
		{give: `env -u VAR sh -c "kill 1"`, wantInner: "kill 1", wantUnwrapped: true},
		// Fix B: env with multiple variable assignments
		{give: `env FOO=1 BAR=2 sh -c "echo test"`, wantInner: "echo test", wantUnwrapped: true},
		// Fix B: env with mixed flags and assignments
		{give: `env -i FOO=1 -u SECRET bash -c "echo ok"`, wantInner: "echo ok", wantUnwrapped: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			inner, unwrapped := unwrapShellWrapper(tt.give)
			assert.Equal(t, tt.wantUnwrapped, unwrapped, "unwrapped")
			assert.Equal(t, tt.wantInner, inner, "inner")
		})
	}
}

func TestIsShellWrapper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{"sh", true},
		{"bash", true},
		{"zsh", true},
		{"dash", true},
		{"python3", false},
		{"node", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isShellWrapper(tt.give))
		})
	}
}

func TestUnwrapShellWrapper_DepthLimitPublic(t *testing.T) {
	t.Parallel()

	// Build a deeply nested shell wrapper with 6+ levels:
	// sh -c "sh -c \"sh -c \\\"sh -c \\\\\\\"sh -c \\\\\\\\\\\\\\\"sh -c \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\"echo hello\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\"\\\\\\\\\\\\\\\"\\\\\\\"\\\"\""
	// Instead of constructing escaped strings, we test that after maxUnwrapDepth
	// (5) levels of recursive unwrap, the function stops.
	//
	// We construct a command that nests 6 levels deep.
	// At depth 5, unwrapShellWrapperAST returns false, so unwrapShellWrapper
	// should return the 5th-level wrapper (not fully unwrapped) or the result
	// at the depth limit.
	//
	// Simple approach: build nesting programmatically.
	// Level 0: "echo hello"
	// Level 1: sh -c "echo hello"
	// Level 2: sh -c 'sh -c "echo hello"'
	// ...
	// For levels > maxUnwrapDepth, unwrap should stop before reaching the innermost.

	// Build from inside out. At each level we wrap with sh -c '...'
	inner := "echo hello"
	for i := 0; i < 6; i++ {
		// Alternate single and double quotes to avoid needing escapes.
		if i%2 == 0 {
			inner = `sh -c "` + inner + `"`
		} else {
			inner = `sh -c '` + inner + `'`
		}
	}

	result, unwrapped := unwrapShellWrapper(inner)

	// The function should unwrap but NOT reach "echo hello" because depth limit
	// is 5 and we have 6 wrapper levels. The result should differ from "echo hello".
	// If depth limiting works, the deepest reachable unwrap stops before "echo hello".
	if unwrapped {
		assert.NotEqual(t, "echo hello", result,
			"6 levels of nesting should hit depth limit before reaching innermost command")
	}
	// Either way, it should not panic or error out.
}

func TestUnwrapShellWrapperAST_DepthLimit(t *testing.T) {
	t.Parallel()

	// Parse a simple shell wrapper command.
	parser := syntax.NewParser()
	f, err := parser.Parse(strings.NewReader(`sh -c "echo hello"`), "")
	assert.NoError(t, err)

	// At depth 4 (under limit of 5), unwrap succeeds.
	inner, ok := unwrapShellWrapperAST(f, 4)
	assert.True(t, ok, "depth 4 should unwrap")
	assert.Equal(t, "echo hello", inner)

	// At depth 5 (at limit), unwrap fails.
	_, ok = unwrapShellWrapperAST(f, 5)
	assert.False(t, ok, "depth 5 should hit limit")

	// At depth 6 (over limit), unwrap fails.
	_, ok = unwrapShellWrapperAST(f, 6)
	assert.False(t, ok, "depth 6 should hit limit")
}

func TestUnwrapShellWrapper_ASTFallback(t *testing.T) {
	t.Parallel()

	// Commands with unmatched quotes fail AST parsing.
	// The string-based fallback should still handle simple patterns.
	// Note: `sh -c "kill 1234` has unmatched quote — AST parse fails,
	// and string fallback also returns false (correct behavior).
	inner, ok := unwrapShellWrapper(`sh -c "kill 1234`)
	assert.False(t, ok, "unmatched quote should not unwrap")
	assert.Equal(t, `sh -c "kill 1234`, inner)
}

func TestLooksLikeEnvAssignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{"FOO=1", true},
		{"BAR=hello", true},
		{"_VAR=value", true},
		{"MY_VAR_2=test", true},
		{"A=", true},         // empty value is valid
		{"./foo=bar", false},  // starts with dot-slash
		{"--flag=val", false}, // starts with dash
		{"=value", false},     // no name before =
		{"noequalssign", false},
		{"", false},
		{"123=bad", false}, // starts with digit
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, looksLikeEnvAssignment(tt.give))
		})
	}
}

func TestStripQuotes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{`"hello world"`, "hello world"},
		{`'hello world'`, "hello world"},
		{`hello world`, "hello world"},
		{`"mismatched'`, `"mismatched'`},
		{`""`, ""},
		{`"a"`, "a"},
		{`a`, "a"},
		{``, ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, stripQuotes(tt.give))
		})
	}
}
