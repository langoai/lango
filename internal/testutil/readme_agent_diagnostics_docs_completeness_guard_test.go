package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedAgentDiagnosticsCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeAgentDiagnosticsDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}

	requiredSnippets := []string{
		"lango agent trace list",
		"lango agent trace show <trace-id>",
		"lango agent graph <session>",
		"lango agent trace metrics",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required agent diagnostics snippet %q", target, snippet)
			}
		}
	}
}

func readmeAgentDiagnosticsDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
