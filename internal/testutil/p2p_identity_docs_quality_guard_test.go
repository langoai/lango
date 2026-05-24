package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestP2PIdentityDocsStayTruthfulAboutCLIOutput(t *testing.T) {
	t.Parallel()

	repoRoot := p2pIdentityDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "features", "p2p-network.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}

	text := string(data)
	forbidden := "The current `lango p2p identity` CLI still focuses on peer/node identity and listen addresses rather than printing the DID directly."
	if strings.Contains(text, forbidden) {
		t.Fatalf("%s contains stale p2p identity CLI wording", target)
	}

	required := "The `lango p2p identity` CLI also prints the active DID directly when available"
	if !strings.Contains(text, required) {
		t.Fatalf("%s must describe that `lango p2p identity` prints the active DID when available", target)
	}

	requiredSummary := "lango p2p identity             # Show local DID, peer identity, and listen addresses"
	if !strings.Contains(text, requiredSummary) {
		t.Fatalf("%s must summarize `lango p2p identity` as including the DID in the CLI command list", target)
	}
}

func p2pIdentityDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
