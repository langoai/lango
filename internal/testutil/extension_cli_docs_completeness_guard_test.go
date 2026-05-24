package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPublicQuickReferencesIncludeImplementedExtensionCommands(t *testing.T) {
	t.Parallel()

	repoRoot := extensionCLIDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}
	requiredSnippets := []string{
		"lango extension inspect <source>",
		"lango extension install <source>",
		"lango extension list",
		"lango extension remove <name>",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required extension quick-reference snippet %q", target, snippet)
			}
		}
	}
}

func extensionCLIDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
