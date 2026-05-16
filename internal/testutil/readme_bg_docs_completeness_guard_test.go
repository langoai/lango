package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedBackgroundCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeBGDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango bg list",
		"lango bg status <id>",
		"lango bg cancel <id>",
		"lango bg result <id>",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README background snippet %q", target, snippet)
		}
	}
}

func readmeBGDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
