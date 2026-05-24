package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestP2PFeatureCommandSummariesStayTruthful(t *testing.T) {
	t.Parallel()

	repoRoot := p2pFeatureCommandDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "features", "p2p-network.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}

	text := string(data)
	forbidden := []string{
		"The CLI commands create ephemeral P2P nodes for one-off operations, independent of the running server:",
		"The `team`, `workspace`, and `git` families below are still guidance surfaces",
		"lango p2p workspace create <name>               # Inspect workspace-create guidance",
		"lango p2p workspace list                        # Inspect runtime-backed workspace listing behavior",
		"lango p2p workspace status <id>                 # Inspect workspace status guidance",
		"lango p2p workspace join <id>                   # Inspect workspace-join guidance",
		"lango p2p workspace leave <id>                  # Inspect workspace-leave guidance",
		"lango p2p git push <workspace-id>               # Create and push git bundle",
		"lango p2p git fetch <workspace-id>              # Fetch and apply git bundle",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			t.Fatalf("%s contains stale P2P git command summary %q", target, needle)
		}
	}

	required := []string{
		"Core inspection commands create ephemeral P2P nodes for one-off operations, independent of the running server. The `workspace` family manages local workspace lifecycle records, while the `team` and `git` families below remain guidance surfaces for server-backed runtime workflows rather than full direct live control:",
		"lango p2p workspace create <name>               # Create a local collaborative workspace",
		"lango p2p workspace list                        # List local collaborative workspaces",
		"lango p2p workspace status <workspace-id>       # Show one local collaborative workspace",
		"lango p2p workspace join <workspace-id>         # Join a local collaborative workspace",
		"lango p2p workspace leave <workspace-id>        # Leave a local collaborative workspace",
		"lango p2p git push <workspace-id>               # Inspect server-backed git bundle push guidance",
		"lango p2p git fetch <workspace-id>              # Inspect server-backed git bundle fetch guidance",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("%s must contain truthful P2P git command summary %q", target, needle)
		}
	}
}

func p2pFeatureCommandDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
