package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompanionConnectivityDocsAvoidStaleDiscoveryClaims(t *testing.T) {
	t.Parallel()

	repoRoot := companionConnectivityDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "architecture", "overview.md"),
		filepath.Join(repoRoot, "docs", "security", "encryption.md"),
		filepath.Join(repoRoot, "openspec", "specs", "companion-discovery", "spec.md"),
	}

	forbidden := []string{
		"_lango-companion._tcp",
		"security.companion.address",
		"auto-discover companion apps on the local network using mDNS",
		"Lango can auto-discover companion apps on the local network using mDNS:",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains stale companion discovery claim %q", target, needle)
			}
		}
	}
}

func companionConnectivityDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
