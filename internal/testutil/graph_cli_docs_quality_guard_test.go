package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGraphCLIDocsDescribeCurrentCommandSurface(t *testing.T) {
	t.Parallel()

	repoRoot := graphCLIDocsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "graph.md")

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
		"lango graph import <file>",
		"table|json",
		"--format json|csv",
		"--force",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required graph docs snippet %q", target, snippet)
		}
	}
}

func graphCLIDocsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
