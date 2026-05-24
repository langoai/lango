package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestP2POnchainExamplesSpecMatchesDiscoveryScriptReality(t *testing.T) {
	t.Parallel()

	repoRoot := p2pOnchainExamplesGuardRepoRoot(t)
	specPath := filepath.Join(repoRoot, "openspec", "specs", "p2p-onchain-examples", "spec.md")

	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}

	text := string(data)
	forbiddenCount := "Six Docker Compose-based integration examples"
	if strings.Contains(text, forbiddenCount) {
		t.Fatalf("%s contains stale example count claim", specPath)
	}

	requiredCount := "Seven Docker Compose-based integration examples"
	if !strings.Contains(text, requiredCount) {
		t.Fatalf("%s must reflect the current seven-example inventory", specPath)
	}

	if strings.Contains(text, "**Tests** (") {
		t.Fatalf("%s contains stale exact test-count claims for evolving example scripts", specPath)
	}

	forbidden := "- mDNS discovery: polling loop (5s interval, up to 60-90s) instead of fixed sleep"
	if strings.Contains(text, forbidden) {
		t.Fatalf("%s contains stale universal polling claim for example discovery scripts", specPath)
	}

	required := "while `p2p-trading` currently uses a fixed `sleep 15` warm-up before peer checks"
	if !strings.Contains(text, required) {
		t.Fatalf("%s must document the current fixed-sleep exception in p2p-trading", specPath)
	}
}

func p2pOnchainExamplesGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
