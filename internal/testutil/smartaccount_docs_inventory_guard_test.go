package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSmartAccountDocsUseCurrentCommandInventory(t *testing.T) {
	t.Parallel()

	repoRoot := smartAccountDocsInventoryRepoRoot(t)
	targets := map[string][]string{
		filepath.Join(repoRoot, "docs", "cli", "smartaccount.md"): {
			"lango account session create",
			"lango account session revoke",
			"lango account module install",
			"lango account policy set",
			"lango account paymaster approve",
		},
		filepath.Join(repoRoot, "docs", "architecture", "project-structure.md"): {
			"session list/create/revoke",
			"module list/install",
			"policy show/set",
			"paymaster status/approve",
		},
		filepath.Join(repoRoot, "README.md"): {
			"session list/create/revoke",
			"module list/install",
			"policy show/set",
			"paymaster status/approve",
		},
	}

	for target, requiredSnippets := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)

		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required smart-account inventory snippet %q", target, snippet)
			}
		}
	}
}

func smartAccountDocsInventoryRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
