package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedConfigProfileCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeConfigProfileDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango config list",
		"lango config create <name>",
		"lango config use <name>",
		"lango config delete <name>",
		"lango config import <file>",
		"lango config export <name>",
		"lango config validate",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README config-profile snippet %q", target, snippet)
		}
	}
}

func readmeConfigProfileDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
