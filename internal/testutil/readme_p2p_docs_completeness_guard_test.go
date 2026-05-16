package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedP2POperatorCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeP2PDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango p2p status",
		"lango p2p peers",
		"lango p2p connect <multiaddr>",
		"lango p2p disconnect <peer-id>",
		"lango p2p firewall list",
		"lango p2p firewall add",
		"lango p2p firewall remove",
		"lango p2p discover",
		"lango p2p identity",
		"lango p2p reputation",
		"lango p2p pricing",
		"lango p2p provenance push <peer-did> <session-key>",
		"lango p2p provenance fetch <peer-did> <session-key>",
		"lango p2p session list",
		"lango p2p session revoke",
		"lango p2p session revoke-all",
		"lango p2p sandbox status",
		"lango p2p sandbox test",
		"lango p2p sandbox cleanup",
		"lango p2p workspace create <name>",
		"lango p2p workspace list",
		"lango p2p workspace status <workspace-id>",
		"lango p2p workspace join <workspace-id>",
		"lango p2p workspace leave <workspace-id>",
		"lango p2p team list",
		"lango p2p team status <id>",
		"lango p2p team disband <id>",
		"lango p2p zkp status",
		"lango p2p zkp circuits",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README P2P snippet %q", target, snippet)
		}
	}
}

func readmeP2PDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
