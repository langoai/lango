package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPublicEconomyQuickReferencesIncludeImplementedEscrowCommands(t *testing.T) {
	t.Parallel()

	repoRoot := economyCLIDocsRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}

	requiredSnippets := []string{
		"lango economy escrow list",
		"lango economy escrow show",
		"lango economy escrow sentinel status",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required economy CLI doc snippet %q", target, snippet)
			}
		}
	}
}

func economyCLIDocsRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
