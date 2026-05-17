package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedP2POperatorCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeP2PDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
	}
	requiredSnippets := []string{
		"lango p2p status",
		"lango p2p peers",
		"lango p2p connect <multiaddr>",
		"lango p2p disconnect <peer-id>",
		"lango p2p firewall list",
		"lango p2p firewall add --peer-did <did>",
		"lango p2p firewall remove <peer-did>",
		"lango p2p discover",
		"lango p2p identity",
		"lango p2p reputation --peer-did <did>",
		"lango p2p pricing",
		"lango p2p provenance push <peer-did> <session-key>",
		"lango p2p provenance fetch <peer-did> <session-key>",
		"lango p2p session list",
		"lango p2p session revoke --peer-did <did>",
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
	staleSnippets := []string{
		"lango p2p firewall add           Add a firewall ACL rule",
		"lango p2p firewall remove        Remove firewall rules for a peer",
		"lango p2p session revoke         Revoke a peer session",
		"`lango p2p firewall add` | Add a firewall ACL rule",
		"`lango p2p firewall remove` | Remove firewall rules for a peer",
		"`lango p2p session revoke` | Revoke a peer session",
		"lango p2p firewall add         # Add a firewall rule",
		"lango p2p session revoke     # Revoke a specific session",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required P2P snippet %q", target, snippet)
			}
		}
		for _, snippet := range staleSnippets {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s contains stale P2P quick-reference snippet %q", target, snippet)
			}
		}
	}

	featureTargets := map[string][]string{
		filepath.Join(repoRoot, "docs", "features", "p2p-network.md"): {
			"lango p2p firewall add --peer-did <did>",
			"lango p2p session revoke --peer-did <did>",
		},
		filepath.Join(repoRoot, "docs", "features", "zkp.md"): {
			"lango p2p session revoke --peer-did <did>",
		},
	}
	for target, snippets := range featureTargets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, snippet := range snippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required P2P snippet %q", target, snippet)
			}
		}
		for _, snippet := range staleSnippets {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s contains stale P2P quick-reference snippet %q", target, snippet)
			}
		}
	}
}

func TestREADMERejectsBareP2PReputationQuickReferenceRow(t *testing.T) {
	t.Parallel()

	target := filepath.Join(readmeP2PDocsGuardRepoRoot(t), "README.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	rowPattern := `(?m)^lango p2p reputation\s+Query peer trust score$`
	if regexp.MustCompile(rowPattern).MatchString(text) {
		t.Fatalf("%s contains stale bare P2P reputation quick-reference row", target)
	}
}

func TestMainSpecsIncludeP2PReputationRequiredPeerDID(t *testing.T) {
	t.Parallel()

	repoRoot := readmeP2PDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "openspec", "specs", "docs-only", "spec.md"),
		filepath.Join(repoRoot, "openspec", "specs", "downstream-docs-sync", "spec.md"),
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		if !strings.Contains(string(data), "lango p2p reputation --peer-did <did>") {
			t.Fatalf("%s is missing required P2P reputation peer DID snippet", target)
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
