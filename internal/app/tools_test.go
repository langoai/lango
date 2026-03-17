package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	execpkg "github.com/langoai/lango/internal/tools/exec"
)

func TestBlockLangoExec_SkillGuards(t *testing.T) {
	tests := []struct {
		give    string
		wantMsg bool
	}{
		{
			give:    "git clone https://github.com/owner/skill-repo",
			wantMsg: true,
		},
		{
			give:    "Git Clone https://github.com/owner/skills",
			wantMsg: true,
		},
		{
			give:    "curl https://example.com/skill.md",
			wantMsg: true,
		},
		{
			give:    "wget https://example.com/skills/SKILL.md",
			wantMsg: true,
		},
		{
			give:    "git clone https://github.com/owner/unrelated-repo",
			wantMsg: false,
		},
		{
			give:    "curl https://example.com/api/data",
			wantMsg: false,
		},
		{
			give:    "ls -la",
			wantMsg: false,
		},
		{
			give:    "lango cron list",
			wantMsg: true,
		},
	}

	auto := map[string]bool{"cron": true, "background": true}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			msg := blockLangoExec(tt.give, auto)
			gotMsg := msg != ""
			assert.Equal(t, tt.wantMsg, gotMsg, "blockLangoExec(%q) returned msg=%q", tt.give, msg)
		})
	}
}

func TestBlockLangoExec_AllSubcommands(t *testing.T) {
	auto := map[string]bool{"cron": true, "background": true, "workflow": true}

	tests := []struct {
		give        string
		wantBlocked bool
		wantContain string // substring expected in the message
	}{
		// Phase 1: subcommands with in-process tool equivalents
		{give: "lango cron list", wantBlocked: true, wantContain: "cron_"},
		{give: "lango bg submit", wantBlocked: true, wantContain: "bg_"},
		{give: "lango background list", wantBlocked: true, wantContain: "bg_"},
		{give: "lango workflow run", wantBlocked: true, wantContain: "workflow_"},
		{give: "lango graph query", wantBlocked: true, wantContain: "graph_"},
		{give: "lango memory list", wantBlocked: true, wantContain: "memory_"},
		{give: "lango p2p status", wantBlocked: true, wantContain: "p2p_"},
		{give: "lango security keyring status", wantBlocked: true, wantContain: "crypto_"},
		{give: "lango payment send", wantBlocked: true, wantContain: "payment_"},
		{give: "lango account deploy", wantBlocked: true, wantContain: "smart_account_deploy"},

		// Phase 2: catch-all for subcommands without in-process equivalents
		{give: "lango config list", wantBlocked: true, wantContain: "passphrase"},
		{give: "lango doctor", wantBlocked: true, wantContain: "passphrase"},
		{give: "lango serve", wantBlocked: true, wantContain: "passphrase"},
		{give: "lango settings", wantBlocked: true, wantContain: "passphrase"},
		{give: "lango onboard", wantBlocked: true, wantContain: "passphrase"},
		{give: "lango agent list", wantBlocked: true, wantContain: "passphrase"},
		{give: "lango", wantBlocked: true, wantContain: "passphrase"},
		{give: "LANGO SECURITY DB-MIGRATE", wantBlocked: true, wantContain: "crypto_"},

		// Not blocked: non-lango commands
		{give: "ls -la", wantBlocked: false},
		{give: "go build ./...", wantBlocked: false},
		{give: "echo lango", wantBlocked: false},
		{give: "cat lango.yaml", wantBlocked: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			msg := blockLangoExec(tt.give, auto)
			gotBlocked := msg != ""
			assert.Equal(t, tt.wantBlocked, gotBlocked, "blockLangoExec(%q) msg=%q", tt.give, msg)
			if tt.wantContain != "" {
				assert.Contains(t, msg, tt.wantContain)
			}
		})
	}
}

func TestBlockLangoExec_DisabledFeature(t *testing.T) {
	// When automation features are disabled, cron/bg/workflow guards
	// should still block but suggest enabling the feature.
	auto := map[string]bool{}

	msg := blockLangoExec("lango cron list", auto)
	require.NotEmpty(t, msg, "expected blocked message for disabled cron")
	assert.Contains(t, msg, "Enable the")

	// Non-automation guards (graph, memory, etc.) should always block
	// regardless of automation flags.
	msg = blockLangoExec("lango graph query", auto)
	require.NotEmpty(t, msg, "expected blocked message for graph")
	assert.NotContains(t, msg, "Enable the", "graph guard should not suggest enabling a feature")
}

func TestBlockProtectedPaths(t *testing.T) {
	guard := execpkg.NewCommandGuard([]string{"~/.lango"})

	tests := []struct {
		give        string
		wantBlocked bool
	}{
		{"sqlite3 ~/.lango/lango.db", true},
		{"cat ~/.lango/keyfile", true},
		{"kill 1", true},
		{"pkill lango", true},
		{"go build ./...", false},
		{"sqlite3 /tmp/test.db", false},
		{"ls -la", false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			msg := blockProtectedPaths(tt.give, guard)
			gotBlocked := msg != ""
			assert.Equal(t, tt.wantBlocked, gotBlocked, "blockProtectedPaths(%q) msg=%q", tt.give, msg)
		})
	}
}

func TestBlockProtectedPaths_NilGuard(t *testing.T) {
	msg := blockProtectedPaths("sqlite3 ~/.lango/lango.db", nil)
	assert.Empty(t, msg, "nil guard should not block anything")
}
