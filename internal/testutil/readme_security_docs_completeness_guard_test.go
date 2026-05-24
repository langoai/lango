package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedSecurityCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeSecurityDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}

	requiredSnippets := []string{
		"lango security status",
		"lango security change-passphrase",
		"lango security migrate-passphrase",
		"lango security secrets list",
		"lango security secrets set <name>",
		"lango security secrets delete <name>",
		"lango security keyring store",
		"lango security keyring clear",
		"lango security keyring status",
		"lango security recovery setup",
		"lango security recovery restore",
		"lango security db-migrate",
		"lango security db-decrypt",
		"lango security kms status",
		"lango security kms test",
		"lango security kms keys",
		"lango security kms wrap",
		"lango security kms detach",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required security snippet %q", target, snippet)
			}
		}
	}
}

func readmeSecurityDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
