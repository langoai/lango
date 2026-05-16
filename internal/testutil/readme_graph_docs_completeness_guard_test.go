package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedGraphCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeGraphDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango graph status",
		"lango graph query",
		"lango graph stats",
		"lango graph clear",
		"lango graph add",
		"lango graph export",
		"lango graph import",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README graph snippet %q", target, snippet)
		}
	}
}

func readmeGraphDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
