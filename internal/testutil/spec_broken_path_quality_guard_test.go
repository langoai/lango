package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMainSpecsAvoidKnownBrokenSinglePathReferences(t *testing.T) {
	t.Parallel()

	repoRoot := specBrokenPathGuardRepoRoot(t)
	checks := []struct {
		path      string
		forbidden []string
	}{
		{
			path: filepath.Join(repoRoot, "openspec", "specs", "shared-types", "spec.md"),
			forbidden: []string{
				"`internal/cli/common/`",
			},
		},
		{
			path: filepath.Join(repoRoot, "openspec", "specs", "skill-runtime-v2", "spec.md"),
			forbidden: []string{
				"`cmd/main.go`",
			},
		},
		{
			path: filepath.Join(repoRoot, "openspec", "specs", "x402-v2", "spec.md"),
			forbidden: []string{
				"`internal/x402/handler.go`",
			},
		},
		{
			path: filepath.Join(repoRoot, "openspec", "specs", "phantom-feature-wiring", "spec.md"),
			forbidden: []string{
				"`internal/companion/discovery.go`",
			},
		},
		{
			path: filepath.Join(repoRoot, "openspec", "specs", "p2p-trading-example", "spec.md"),
			forbidden: []string{
				"`contracts/MockUSDC.sol`",
				"`scripts/test-p2p-trading.sh`",
				"`docker-entrypoint-p2p.sh`",
			},
		},
	}

	for _, check := range checks {
		data, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		text := string(data)
		for _, needle := range check.forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains stale broken path %q", check.path, needle)
			}
		}
	}
}

func specBrokenPathGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
