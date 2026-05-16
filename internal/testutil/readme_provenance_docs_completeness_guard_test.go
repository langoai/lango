package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedProvenanceCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeProvenanceDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "openspec", "specs", "docs-only", "spec.md"),
	}

	requiredSnippets := []string{
		"lango provenance status",
		"lango provenance checkpoint list",
		"lango provenance checkpoint create",
		"lango provenance checkpoint show <id>",
		"lango provenance session tree",
		"lango provenance session list",
		"lango provenance attribution show <session>",
		"lango provenance attribution report",
		"lango provenance bundle export",
		"lango provenance bundle import",
	}

	for _, target := range targets {
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

func readmeProvenanceDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
