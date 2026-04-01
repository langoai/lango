package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		// Supported: no quotes around inner command
		{give: `sh -c echo hello`, wantInner: "echo hello", wantUnwrapped: true},
		// Supported: zsh
		{give: `zsh -c "ls -la"`, wantInner: "ls -la", wantUnwrapped: true},
		// Supported: dash
		{give: `dash -c 'go build ./...'`, wantInner: "go build ./...", wantUnwrapped: true},
		// Not a shell wrapper: python3 -c
		{give: `python3 -c "print('hi')"`, wantInner: `python3 -c "print('hi')"`, wantUnwrapped: false},
		// Not a shell wrapper: sh without -c
		{give: `sh script.sh`, wantInner: `sh script.sh`, wantUnwrapped: false},
		// Not supported: login shell flag -lc
		{give: `sh -lc "cmd"`, wantInner: `sh -lc "cmd"`, wantUnwrapped: false},
		// Not supported: env wrapper
		{give: `/usr/bin/env sh -c "cmd"`, wantInner: `/usr/bin/env sh -c "cmd"`, wantUnwrapped: false},
		// Not supported: env prefix
		{give: `env bash -c "cmd"`, wantInner: `env bash -c "cmd"`, wantUnwrapped: false},
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
