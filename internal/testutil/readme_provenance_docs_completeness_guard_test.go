package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedProvenanceCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeProvenanceDocsGuardRepoRoot(t)
	targetRequirements := map[string][]string{
		filepath.Join(repoRoot, "README.md"): {
			"lango provenance status",
			"lango provenance checkpoint list --run <id>",
			"lango provenance checkpoint create <label> --run <id>",
			"lango provenance checkpoint show <id>",
			"lango provenance session tree <session-key>",
			"lango provenance session list",
			"lango provenance attribution show <session-key>",
			"lango provenance attribution report <session-key>",
			"lango provenance bundle export <session-key>",
			"lango provenance bundle import <file>",
		},
		filepath.Join(repoRoot, "openspec", "specs", "docs-only", "spec.md"): {
			"lango provenance status",
			"lango provenance checkpoint list --run <id>",
			"lango provenance checkpoint create <label> --run <id>",
			"lango provenance checkpoint show <id>",
			"lango provenance session tree <session-key>",
			"lango provenance session list",
			"lango provenance attribution show <session-key>",
			"lango provenance attribution report <session-key>",
			"lango provenance bundle export <session-key>",
			"lango provenance bundle import <file>",
		},
		filepath.Join(repoRoot, "openspec", "specs", "downstream-docs-sync", "spec.md"): {
			"lango provenance checkpoint list --run <id>",
			"lango provenance checkpoint create <label> --run <id>",
			"lango provenance checkpoint show <id>",
			"lango provenance session tree <session-key>",
			"lango provenance session list",
			"lango provenance attribution show <session-key>",
			"lango provenance attribution report <session-key>",
			"lango provenance bundle export <session-key>",
			"lango provenance bundle import <file>",
		},
	}

	for target, requiredSnippets := range targetRequirements {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required provenance snippet %q", target, snippet)
			}
		}
	}
}

func TestREADMERejectsBareProvenanceQuickReferenceRows(t *testing.T) {
	t.Parallel()

	target := filepath.Join(readmeProvenanceDocsGuardRepoRoot(t), "README.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	forbiddenRows := []string{
		`^lango provenance checkpoint list\s+List checkpoints$`,
		`^lango provenance checkpoint create\s+Create a manual checkpoint$`,
		`^lango provenance session tree\s+Show session hierarchy tree$`,
		`^lango provenance attribution show <session>\s+Show attribution data for a session$`,
		`^lango provenance attribution report\s+Generate attribution report$`,
		`^lango provenance bundle export\s+Export a signed provenance bundle$`,
		`^lango provenance bundle import\s+Import a signed provenance bundle$`,
	}
	for _, rowPattern := range forbiddenRows {
		if regexp.MustCompile(`(?m)` + rowPattern).MatchString(text) {
			t.Fatalf("%s contains stale bare provenance quick-reference row matching %q", target, rowPattern)
		}
	}
}

func readmeProvenanceDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
