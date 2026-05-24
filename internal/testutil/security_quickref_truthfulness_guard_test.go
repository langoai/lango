package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSecurityQuickReferencesDistinguishCanonicalAndDeprecatedPassphraseCommands(t *testing.T) {
	t.Parallel()

	repoRoot := securityQuickRefTruthfulnessRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}

	requiredSnippets := []string{
		"lango security change-passphrase",
		"without re-encrypting all data",
		"lango security migrate-passphrase",
		"[DEPRECATED] Legacy full re-encryption passphrase migration",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required security quick-reference truthfulness snippet %q", target, snippet)
			}
		}
	}
}

func securityQuickRefTruthfulnessRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
