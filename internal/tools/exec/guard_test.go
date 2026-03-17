package exec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandGuard_CheckCommand(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	langoDir := filepath.Join(home, ".lango")

	guard := NewCommandGuard([]string{"~/.lango"})

	tests := []struct {
		give        string
		wantBlocked bool
		wantContain string
	}{
		// Blocked: direct DB access
		{
			give:        "sqlite3 ~/.lango/lango.db",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		{
			give:        "sqlite3 " + langoDir + "/lango.db",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		// Blocked: reading protected files
		{
			give:        "cat ~/.lango/keyfile",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		{
			give:        "python3 -c 'import sqlite3; c=sqlite3.connect(\"" + langoDir + "/lango.db\")'",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		// Blocked: $HOME variant
		{
			give:        "cat $HOME/.lango/lango.db",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		// Blocked: ${HOME} variant
		{
			give:        "sqlite3 ${HOME}/.lango/graph.db",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		// Blocked: pipe chains accessing protected path
		{
			give:        "cat ~/.lango/keyfile | base64",
			wantBlocked: true,
			wantContain: "protected data path",
		},
		// Blocked: process management
		{
			give:        "kill 1",
			wantBlocked: true,
			wantContain: "process management",
		},
		{
			give:        "pkill lango",
			wantBlocked: true,
			wantContain: "process management",
		},
		{
			give:        "killall lango",
			wantBlocked: true,
			wantContain: "process management",
		},
		{
			give:        "KILL -9 1234",
			wantBlocked: true,
			wantContain: "process management",
		},
		// Allowed: normal commands
		{
			give:        "go build ./...",
			wantBlocked: false,
		},
		{
			give:        "ls -la",
			wantBlocked: false,
		},
		{
			give:        "grep kill log.txt",
			wantBlocked: false,
		},
		{
			give:        "echo 'kill process'",
			wantBlocked: false,
		},
		{
			give:        "sqlite3 /tmp/test.db",
			wantBlocked: false,
		},
		{
			give:        "cat /etc/hosts",
			wantBlocked: false,
		},
		{
			give:        "python3 script.py",
			wantBlocked: false,
		},
		// Allowed: word "kill" in arguments (not as verb)
		{
			give:        "grep -r kill src/",
			wantBlocked: false,
		},
		{
			give:        "echo killswitch",
			wantBlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			blocked, reason := guard.CheckCommand(tt.give)
			assert.Equal(t, tt.wantBlocked, blocked, "CheckCommand(%q) blocked=%v reason=%q", tt.give, blocked, reason)
			if tt.wantContain != "" {
				assert.Contains(t, reason, tt.wantContain)
			}
		})
	}
}

func TestCommandGuard_MultipleProtectedPaths(t *testing.T) {
	t.Parallel()

	guard := NewCommandGuard([]string{"~/.lango", "/var/secrets"})

	blocked, _ := guard.CheckCommand("cat /var/secrets/api.key")
	assert.True(t, blocked, "should block access to /var/secrets")

	blocked, _ = guard.CheckCommand("cat /tmp/safe.txt")
	assert.False(t, blocked, "should allow access to /tmp")
}

func TestCommandGuard_EmptyCommand(t *testing.T) {
	t.Parallel()

	guard := NewCommandGuard([]string{"~/.lango"})
	blocked, _ := guard.CheckCommand("")
	assert.False(t, blocked, "empty command should not be blocked")
}

func TestCommandGuard_SubCommands(t *testing.T) {
	t.Parallel()

	guard := NewCommandGuard([]string{"~/.lango"})

	tests := []struct {
		give        string
		wantBlocked bool
	}{
		{"echo hello && cat ~/.lango/keyfile", true},
		{"ls; sqlite3 ~/.lango/lango.db", true},
		{"echo hello && echo world", false},
		{"cat /tmp/a | grep test", false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			blocked, _ := guard.CheckCommand(tt.give)
			assert.Equal(t, tt.wantBlocked, blocked, "CheckCommand(%q)", tt.give)
		})
	}
}

func TestExtractVerb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{"kill 1", "kill"},
		{"/usr/bin/kill -9 1234", "kill"},
		{"  PKILL lango", "pkill"},
		{"ls -la", "ls"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := extractVerb(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
