package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedSmartAccountCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeSmartAccountDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "index.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango account info",
		"lango account deploy",
		"lango account session list",
		"lango account session create",
		"lango account session revoke",
		"lango account module list",
		"lango account module install",
		"lango account policy show",
		"lango account policy set",
		"lango account paymaster status",
		"lango account paymaster approve",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README smart account snippet %q", target, snippet)
		}
	}
}

func readmeSmartAccountDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
