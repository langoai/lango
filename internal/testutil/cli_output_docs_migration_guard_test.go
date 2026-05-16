package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestMigratedCLIDocsUseCurrentOutputFormatContracts(t *testing.T) {
	t.Parallel()

	repoRoot := cliOutputDocsMigrationGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "a2a.md"),
		filepath.Join(repoRoot, "docs", "cli", "approval.md"),
		filepath.Join(repoRoot, "docs", "cli", "agent-memory.md"),
		filepath.Join(repoRoot, "docs", "cli", "contract.md"),
		filepath.Join(repoRoot, "docs", "cli", "core.md"),
		filepath.Join(repoRoot, "docs", "cli", "agent-memory.md"),
		filepath.Join(repoRoot, "docs", "cli", "learning.md"),
		filepath.Join(repoRoot, "docs", "cli", "librarian.md"),
		filepath.Join(repoRoot, "docs", "cli", "payment.md"),
		filepath.Join(repoRoot, "docs", "cli", "run.md"),
		filepath.Join(repoRoot, "docs", "cli", "sandbox.md"),
		filepath.Join(repoRoot, "docs", "cli", "security.md"),
		filepath.Join(repoRoot, "docs", "cli", "automation.md"),
		filepath.Join(repoRoot, "docs", "features", "knowledge-graph.md"),
		filepath.Join(repoRoot, "docs", "payments", "usdc.md"),
		filepath.Join(repoRoot, "docs", "security", "approval-cli.md"),
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-a2a-management", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-agent-memory", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-agent-tools-hooks", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-approval-dashboard", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-doctor", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-help-text", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-graph-extended", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-graph-management", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-health-check", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-learning-inspection", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-librarian-monitoring", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-memory-management", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-payment-management", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-secrets-management", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-security-status", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-workflow-validate", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cli-x402-config", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "cloud-kms", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "contract-interaction", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "run-ledger", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "security-docs-sync", "spec.md"),
	}

	forbiddenPatterns := []struct {
		re     *regexp.Regexp
		reason string
	}{
		{
			re:     regexp.MustCompile(`(?s)lango a2a card[\s\S]{0,400}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale a2a card boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango a2a check[\s\S]{0,400}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale a2a check boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango approval status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale approval boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango agent status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale agent status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango agent list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale agent list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango agent tools[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale agent tools boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango agent hooks[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale agent hooks boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango doctor[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output results as JSON \|`),
			reason: "stale doctor boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango sandbox status[\s\S]{0,300}?\| \x60--json\x60 \| \x60bool\x60 \| Output results as JSON \|`),
			reason: "stale sandbox status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango sandbox test[\s\S]{0,300}?\| \x60--json\x60 \| \x60bool\x60 \| Output results as JSON \|`),
			reason: "stale sandbox test boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango graph status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale graph status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango graph query[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale graph query boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango graph stats[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale graph stats boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango graph add[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output the added triple as JSON \|`),
			reason: "stale graph add boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?m)lango graph query --subject .* --json`),
			reason: "stale graph query example using --json",
		},
		{
			re:     regexp.MustCompile(`(?s)lango graph import[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output the import summary as JSON \|`),
			reason: "stale graph import boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango learning status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale learning status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango learning history[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale learning history boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango librarian status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale librarian status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango librarian inquiries[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale librarian inquiries boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango memory list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale memory list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango memory status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale memory status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango memory agents[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale memory agents boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango memory agent <name>[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale memory agent boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango payment balance[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale payment balance boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango payment history[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale payment history boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango payment limits[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale payment limits boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango payment info[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale payment info boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango payment send[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale payment send boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango payment x402[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale payment x402 boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango workflow validate[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output results as JSON \|`),
			reason: "stale workflow validate boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango run list[\s\S]{0,300}?\| \x60--json\x60 \| \x60bool\x60 \| Output results as JSON \|`),
			reason: "stale run list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango run status[\s\S]{0,300}?\| \x60--json\x60 \| \x60bool\x60 \| Output results as JSON \|`),
			reason: "stale run status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango run journal <run-id>[\s\S]{0,300}?\| \x60--json\x60 \| \x60bool\x60 \| Output results as JSON \|`),
			reason: "stale run journal boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango security status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale security status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango security keyring status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale security keyring status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango security kms status[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale security kms status boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango security kms keys[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale security kms keys boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?s)lango security secrets list[\s\S]{0,300}?\| \x60--json\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale security secrets list boolean --json docs",
		},
		{
			re:     regexp.MustCompile(`(?m)lango (?:a2a (?:card|check)|agent (?:status|list|tools|hooks)|approval status|doctor|graph (?:status|query|stats|add|import)|learning (?:status|history)|librarian (?:status|inquiries)|memory (?:list|status|agents|agent)|payment (?:balance|history|limits|info|send|x402)|security (?:status|keyring status|kms status|kms keys|secrets list)|workflow validate [^\n]*|run (?:list|status|journal)[^\n]*)\s+.*--json`),
			reason: "stale --json example for migrated command family",
		},
		{
			re:     regexp.MustCompile("(?m)Use `--json` to inspect"),
			reason: "stale prose telling users to use --json",
		},
		{
			re:     regexp.MustCompile("(?m)All graph CLI commands support `--json`"),
			reason: "stale graph prose using --json",
		},
		{
			re:     regexp.MustCompile(`(?m)lango payment balance --json`),
			reason: "stale payment balance example using --json",
		},
		{
			re:     regexp.MustCompile(`(?m)Use --json for machine-readable output\.`),
			reason: "stale doctor prose using --json",
		},
		{
			re:     regexp.MustCompile(`(?m)lango doctor --json`),
			reason: "stale doctor example using --json",
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

func cliOutputDocsMigrationGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
