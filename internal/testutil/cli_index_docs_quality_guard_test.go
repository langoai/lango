package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIIndexIncludesImplementedOperatorCommands(t *testing.T) {
	t.Parallel()

	indexPath := filepath.Join(cliIndexDocsRepoRoot(t), "docs", "cli", "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"`lango security kms wrap`",
		"`lango security kms detach`",
		"`lango p2p workspace create <name>`",
		"`lango p2p workspace list`",
		"`lango p2p workspace status <workspace-id>`",
		"`lango p2p workspace join <workspace-id>`",
		"`lango p2p workspace leave <workspace-id>`",
		"`lango p2p provenance push <peer-did> <session-key>`",
		"`lango p2p provenance fetch <peer-did> <session-key>`",
		"`lango p2p reputation --peer-did <did>`",
		"`lango provenance checkpoint list --run <id>`",
		"`lango provenance checkpoint create <label> --run <id>`",
		"`lango provenance session tree <session-key>`",
		"`lango provenance attribution show <session-key>`",
		"`lango provenance attribution report <session-key>`",
		"`lango provenance bundle export <session-key>`",
		"`lango provenance bundle import <file>`",
		"| `lango p2p team list` | Describe how to inspect active P2P teams |",
		"| `lango p2p team status <id>` | Describe how to inspect runtime-backed team status |",
		"| `lango p2p team disband <id>` | Describe how to disband a runtime-backed team |",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required CLI index snippet %q", indexPath, snippet)
		}
	}
}

func TestCLIIndexRejectsBareQuickReferenceRowsWithRequiredOperands(t *testing.T) {
	t.Parallel()

	indexPath := filepath.Join(cliIndexDocsRepoRoot(t), "docs", "cli", "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}
	text := string(data)

	forbiddenSnippets := []string{
		"| `lango p2p reputation` | Query peer trust score |",
		"| `lango provenance checkpoint list` | List checkpoints |",
		"| `lango provenance checkpoint create` | Create a manual checkpoint |",
		"| `lango provenance session tree` | Show session hierarchy tree |",
		"| `lango provenance attribution show <session>` | Show attribution data for a session |",
		"| `lango provenance attribution report` | Generate attribution report |",
		"| `lango provenance bundle export` | Export a signed provenance bundle |",
		"| `lango provenance bundle import` | Import a signed provenance bundle |",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(text, snippet) {
			t.Fatalf("%s contains stale bare quick-reference row %q", indexPath, snippet)
		}
	}
}

func TestCLIIndexAvoidsStaleP2PTeamLiveControlSummaries(t *testing.T) {
	t.Parallel()

	indexPath := filepath.Join(cliIndexDocsRepoRoot(t), "docs", "cli", "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}
	text := string(data)

	forbiddenSnippets := []string{
		"| `lango p2p team list` | List active P2P teams |",
		"| `lango p2p team status <id>` | Show team details and member status |",
		"| `lango p2p team disband <id>` | Disband an active team |",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(text, snippet) {
			t.Fatalf("%s contains stale live-control P2P team summary %q", indexPath, snippet)
		}
	}
}

func TestCLIIndexHasNoProseEmbeddedInsideAgentMemoryTable(t *testing.T) {
	t.Parallel()

	indexPath := filepath.Join(cliIndexDocsRepoRoot(t), "docs", "cli", "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}
	text := string(data)

	badSnippet := "| `lango memory agents` | List agents with persistent memory |\n\n`lango agent status` writes through"
	if strings.Contains(text, badSnippet) {
		t.Fatalf("%s contains prose embedded inside the Agent & Memory command table", indexPath)
	}
}

func cliIndexDocsRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
