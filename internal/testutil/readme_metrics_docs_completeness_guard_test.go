package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedMetricsCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeMetricsDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}

	requiredSnippets := []string{
		"lango metrics",
		"lango metrics sessions",
		"lango metrics tools",
		"lango metrics agents",
		"lango metrics policy",
		"lango metrics history",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required metrics snippet %q", target, snippet)
			}
		}
	}
}

func readmeMetricsDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
