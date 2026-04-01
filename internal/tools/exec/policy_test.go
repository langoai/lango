package exec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEvaluator(t *testing.T) *PolicyEvaluator {
	t.Helper()
	guard := NewCommandGuard([]string{"~/.lango"})
	classifier := func(cmd string) (string, ReasonCode) {
		return testClassifyLangoExec(cmd)
	}
	return NewPolicyEvaluator(guard, classifier, nil,
		WithCatastrophicPatterns([]string{
			"rm -rf /",
			"mkfs.",
			"dd if=/dev/zero",
			":(){ :|:& };:",
			"> /dev/sda",
			"chmod -R 777 /",
			"dd if=/dev/random",
			"mv / ",
			"> /dev/null 2>&1 &",
		}))
}

// testClassifyLangoExec mirrors the real classifyLangoExec logic for testing.
func testClassifyLangoExec(cmd string) (string, ReasonCode) {
	lower := strings.ToLower(strings.TrimSpace(cmd))

	// Lango CLI commands.
	if strings.HasPrefix(lower, "lango ") || lower == "lango" {
		return "blocked: lango CLI requires passphrase", ReasonLangoCLI
	}

	// Skill import redirects.
	if strings.HasPrefix(lower, "git clone") && strings.Contains(lower, "skill") {
		return "use import_skill tool instead", ReasonSkillImport
	}
	if (strings.HasPrefix(lower, "curl ") || strings.HasPrefix(lower, "wget ")) &&
		strings.Contains(lower, "skill") {
		return "use import_skill tool instead", ReasonSkillImport
	}

	return "", ReasonNone
}

func TestPolicyEvaluator_Evaluate(t *testing.T) {
	t.Parallel()

	pe := newTestEvaluator(t)

	tests := []struct {
		give       string
		wantVerdict Verdict
		wantReason  ReasonCode
	}{
		// Block: shell wrapper bypass for kill verb
		{give: `sh -c "kill 1234"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: `bash -c "pkill lango"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: `/bin/sh -c "killall node"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		// Block: shell wrapper bypass for lango CLI
		{give: `bash -c "lango security check"`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		{give: `sh -c "lango cron list"`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// Block: shell wrapper bypass for skill import
		{give: `bash -c "git clone https://github.com/org/repo skill-name"`, wantVerdict: VerdictBlock, wantReason: ReasonSkillImport},
		// Block: direct kill verb (existing behavior preserved)
		{give: "kill 1234", wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: "pkill lango", wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: "killall node", wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		// Block: direct lango CLI (existing behavior preserved)
		{give: "lango cron list", wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		{give: "lango security check", wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// Block: direct protected path (existing behavior preserved)
		{give: "sqlite3 ~/.lango/lango.db", wantVerdict: VerdictBlock, wantReason: ReasonProtectedPath},
		{give: "cat $HOME/.lango/keyfile", wantVerdict: VerdictBlock, wantReason: ReasonProtectedPath},
		// Observe: opaque command substitution
		{give: "ls $(cat /etc/passwd)", wantVerdict: VerdictObserve, wantReason: ReasonCmdSubstitution},
		{give: "echo `whoami`", wantVerdict: VerdictObserve, wantReason: ReasonCmdSubstitution},
		// Block: catastrophic pattern (step 4, before opaque detection)
		{give: `rm -rf /`, wantVerdict: VerdictBlock, wantReason: ReasonCatastrophicPattern},
		{give: `eval "rm -rf /"`, wantVerdict: VerdictBlock, wantReason: ReasonCatastrophicPattern},
		{give: `echo $(mkfs.ext4 /dev/sda)`, wantVerdict: VerdictBlock, wantReason: ReasonCatastrophicPattern},
		{give: `dd if=/dev/zero of=/dev/sda`, wantVerdict: VerdictBlock, wantReason: ReasonCatastrophicPattern},
		// Observe: eval verb (non-catastrophic)
		{give: `eval "echo hello"`, wantVerdict: VerdictObserve, wantReason: ReasonEvalVerb},
		// Observe: unsafe variable expansion
		{give: "echo $SECRET_TOKEN", wantVerdict: VerdictObserve, wantReason: ReasonUnsafeVarExpand},
		// Allow: clean commands
		{give: "go build ./...", wantVerdict: VerdictAllow, wantReason: ReasonNone},
		{give: "ls -la", wantVerdict: VerdictAllow, wantReason: ReasonNone},
		{give: "grep kill log.txt", wantVerdict: VerdictAllow, wantReason: ReasonNone},
		// Allow: clean command through shell wrapper
		{give: `sh -c "go build ./..."`, wantVerdict: VerdictAllow, wantReason: ReasonNone},
		{give: `bash -c "echo hello"`, wantVerdict: VerdictAllow, wantReason: ReasonNone},
		// P1 fix: positional args bypass
		{give: `bash -c "kill 1234" ignored`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: `sh -c "lango cron" myname`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// P2: login shell flag -lc blocks kill verb
		{give: `sh -lc "kill 1234"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: `bash -lc "lango security check"`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// P2: env wrapper blocks kill verb
		{give: `/usr/bin/env sh -c "kill 1234"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: `env bash -c "lango cron list"`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// P2: env wrapper allows clean command
		{give: `env sh -c "echo hello"`, wantVerdict: VerdictAllow, wantReason: ReasonNone},
		// P2: nested wrapper blocks inner dangerous command
		{give: `sh -c "bash -c \"kill 9999\""`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		{give: `bash -c 'sh -c "lango security check"'`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// P2: nested wrapper allows clean inner command
		{give: `sh -c "bash -c \"echo safe\""`, wantVerdict: VerdictAllow, wantReason: ReasonNone},
		// P2: env + login shell combination
		{give: `env sh -lc "kill 1234"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		// Fix B: env with variable assignment — kill verb blocked
		{give: `env FOO=1 sh -c "kill 1234"`, wantVerdict: VerdictBlock, wantReason: ReasonKillVerb},
		// Fix B: env -i flag — lango CLI blocked
		{give: `env -i bash -c "lango security"`, wantVerdict: VerdictBlock, wantReason: ReasonLangoCLI},
		// Fix B: env -u flag — clean command allowed
		{give: `env -u SECRET sh -c "echo hi"`, wantVerdict: VerdictAllow, wantReason: ReasonNone},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			d := pe.Evaluate(tt.give)
			assert.Equal(t, tt.wantVerdict, d.Verdict, "verdict for %q", tt.give)
			assert.Equal(t, tt.wantReason, d.Reason, "reason for %q", tt.give)
		})
	}
}

func TestPolicyEvaluator_Evaluate_UnwrappedField(t *testing.T) {
	t.Parallel()

	pe := newTestEvaluator(t)

	// When a shell wrapper is detected, Unwrapped should contain the inner command.
	d := pe.Evaluate(`sh -c "go build ./..."`)
	assert.Equal(t, "go build ./...", d.Unwrapped)
	assert.Equal(t, `sh -c "go build ./..."`, d.Command)

	// When no wrapper, Unwrapped should be empty.
	d = pe.Evaluate("go build ./...")
	assert.Equal(t, "", d.Unwrapped)
	assert.Equal(t, "go build ./...", d.Command)
}

func TestPolicyEvaluator_UserConfiguredBlockedPattern(t *testing.T) {
	t.Parallel()

	guard := NewCommandGuard([]string{"~/.lango"})
	classifier := func(cmd string) (string, ReasonCode) { return "", ReasonNone }
	pe := NewPolicyEvaluator(guard, classifier, nil,
		WithCatastrophicPatterns([]string{"drop table", "truncate table"}))

	// User-configured pattern blocks.
	d := pe.Evaluate(`exec "drop table users"`)
	assert.Equal(t, VerdictBlock, d.Verdict)
	assert.Equal(t, ReasonCatastrophicPattern, d.Reason)

	d = pe.Evaluate("truncate table sessions")
	assert.Equal(t, VerdictBlock, d.Verdict)
	assert.Equal(t, ReasonCatastrophicPattern, d.Reason)

	// Non-matching command passes through.
	d = pe.Evaluate("select * from users")
	assert.Equal(t, VerdictAllow, d.Verdict)
}

func TestPolicyEvaluator_NoCatastrophicPatternsPassesAll(t *testing.T) {
	t.Parallel()

	guard := NewCommandGuard([]string{"~/.lango"})
	classifier := func(cmd string) (string, ReasonCode) { return "", ReasonNone }
	// No WithCatastrophicPatterns option → empty slice → step 4 passes everything.
	pe := NewPolicyEvaluator(guard, classifier, nil)

	// Without catastrophic patterns, rm -rf / is allowed (not blocked by step 4).
	d := pe.Evaluate("rm -rf /")
	assert.Equal(t, VerdictAllow, d.Verdict)
	assert.Equal(t, ReasonNone, d.Reason)
}

func TestNewPolicyEvaluator_SafeVarsInitialized(t *testing.T) {
	t.Parallel()

	pe := newTestEvaluator(t)
	require.Len(t, pe.safeVars, 10)

	expectedVars := []string{"HOME", "PATH", "USER", "PWD", "SHELL", "TERM", "LANG", "LC_ALL", "LC_CTYPE", "TMPDIR"}
	for _, v := range expectedVars {
		_, ok := pe.safeVars[v]
		assert.True(t, ok, "safe var %q should be present", v)
	}
}
