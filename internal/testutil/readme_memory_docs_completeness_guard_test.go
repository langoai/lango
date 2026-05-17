package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedMemoryCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeMemoryDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}
	requiredSnippets := []string{
		"lango memory list",
		"lango memory status",
		"lango memory clear <session-key>",
		"lango memory agents",
		"lango memory agent <name>",
	}
	staleSnippets := []string{
		"lango memory clear              Clear all memory entries for a session",
		"`lango memory clear` | Clear all memory entries for a session",
		"`lango memory clear` to manage observation entries",
		"lango memory clear\n",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required memory snippet %q", target, snippet)
			}
		}
		for _, snippet := range staleSnippets {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s contains stale memory quick-reference snippet %q", target, snippet)
			}
		}
	}

	featureDoc := filepath.Join(repoRoot, "docs", "features", "observational-memory.md")
	data, err := os.ReadFile(featureDoc)
	if err != nil {
		t.Fatalf("read %s: %v", featureDoc, err)
	}
	text := string(data)
	if !strings.Contains(text, "lango memory clear <session-key>") {
		t.Fatalf("%s is missing required memory clear session key snippet", featureDoc)
	}
	for _, snippet := range staleSnippets {
		if strings.Contains(text, snippet) {
			t.Fatalf("%s contains stale memory quick-reference snippet %q", featureDoc, snippet)
		}
	}
}

func readmeMemoryDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
