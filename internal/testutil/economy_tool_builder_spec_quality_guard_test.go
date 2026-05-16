package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEconomyToolBuilderSpecsAvoidDeletedAppToolPathClaims(t *testing.T) {
	t.Parallel()

	repoRoot := economyToolBuilderSpecGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "openspec", "specs", "domain-tool-builders", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "escrow-sentinel", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "onchain-escrow", "delta.md"),
	}

	forbidden := []string{
		"internal/app/tools_sentinel.go",
		"tools_economy.go",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains stale deleted tool-builder path %q", target, needle)
			}
		}
	}
}

func economyToolBuilderSpecGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
