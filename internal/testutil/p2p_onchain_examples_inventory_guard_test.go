package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestP2POnchainExamplesSpecListsAllShippedExamples(t *testing.T) {
	t.Parallel()

	repoRoot := p2pOnchainExamplesInventoryGuardRepoRoot(t)
	specPath := filepath.Join(repoRoot, "openspec", "specs", "p2p-onchain-examples", "spec.md")

	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	text := string(data)

	expectedHeadings := []string{
		"### 1. discovery-and-handshake (Beginner)",
		"### 2. smart-account-basics (Beginner/Intermediate)",
		"### 3. firewall-and-reputation (Intermediate)",
		"### 4. p2p-trading (Intermediate)",
		"### 5. paid-tool-marketplace (Intermediate/Advanced)",
		"### 6. escrow-milestones (Advanced)",
		"### 7. team-workspace (Advanced)",
	}

	for _, heading := range expectedHeadings {
		if !strings.Contains(text, heading) {
			t.Fatalf("%s is missing shipped example heading %q", specPath, heading)
		}
	}
}

func p2pOnchainExamplesInventoryGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
