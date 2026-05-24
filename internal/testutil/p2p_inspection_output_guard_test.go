package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestP2PInspectionSubsetAvoidsBooleanJSONFlags(t *testing.T) {
	t.Parallel()

	repoRoot := p2pInspectionGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "internal", "cli", "p2p", "status.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "peers.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "identity.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "discover.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "firewall.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "git.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "pricing.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "reputation.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "session.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "team.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "workspace.go"),
		filepath.Join(repoRoot, "internal", "cli", "p2p", "zkp.go"),
	}

	jsonFlagDecl := regexp.MustCompile(`BoolVarP?\([^)\n]*"json"`)

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		if jsonFlagDecl.Find(data) != nil {
			t.Fatalf("%s reintroduces a boolean --json flag declaration", target)
		}
	}
}

func TestP2PInspectionDocsUseCurrentOutputFormatContracts(t *testing.T) {
	t.Parallel()

	repoRoot := p2pInspectionGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "p2p.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-p2p-management", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "p2p-identity", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "p2p-pricing-cli", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "p2p-reputation-cli", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-p2p-teams", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-zkp-inspection", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "session-invalidation", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "security-docs-sync", "spec.md"),
	}

	forbiddenPatterns := []struct {
		re     *regexp.Regexp
		reason string
	}{
		{
			re:     regexp.MustCompile(`(?s)lango p2p status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p peers[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p peers boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p identity[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p identity boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p discover[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p discover boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p firewall list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p firewall boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p team list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p team list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p team status <team-id>[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p team status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p zkp status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p zkp status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p zkp circuits[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p zkp circuits boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p session list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p session list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p pricing[\s\S]{0,300}?\| \x60--json\x60 \| Output as JSON \|`),
			reason: "stale p2p pricing boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p reputation[\s\S]{0,300}?\| \x60--json\x60 \| Output as JSON \|`),
			reason: "stale p2p reputation boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p workspace create <name>[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p workspace create boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p workspace list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p workspace list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p workspace status <id>[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p workspace status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango p2p git log <workspace-id>[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale p2p git log boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?m)lango p2p (?:status|peers|identity|discover|firewall list|pricing|reputation|session list|team list|team status|zkp status|zkp circuits|workspace create|workspace list|workspace status|git log)\b[^\n]*--json`),
			reason: "stale p2p family --json example",
		},
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		for _, pattern := range forbiddenPatterns {
			if pattern.re.Find(data) != nil {
				t.Fatalf("%s contains %s", target, pattern.reason)
			}
		}
	}
}

func p2pInspectionGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
