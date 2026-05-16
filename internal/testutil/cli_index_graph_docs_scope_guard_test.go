package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIIndexDelegatesGraphCommandsToDedicatedSection(t *testing.T) {
	t.Parallel()

	repoRoot := cliIndexGraphDocsScopeGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "index.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"### Agent & Memory",
		"Graph CLI Reference](graph.md)",
		"### Graph Store",
		"lango graph status",
		"lango graph query",
		"lango graph stats",
		"lango graph clear",
		"lango graph add",
		"lango graph export",
		"lango graph import",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required graph-scope snippet %q", target, snippet)
		}
	}

	agentMemoryStart := strings.Index(text, "### Agent & Memory")
	graphStart := strings.Index(text, "### Graph Store")
	alertsStart := strings.Index(text, "### Alerts")
	if agentMemoryStart == -1 || graphStart == -1 || alertsStart == -1 || !(agentMemoryStart < graphStart && graphStart < alertsStart) {
		t.Fatalf("%s is missing the expected Agent & Memory -> Graph Store -> Alerts section ordering", target)
	}

	agentMemorySection := text[agentMemoryStart:graphStart]
	forbiddenAgentMemorySnippets := []string{
		"lango graph status",
		"lango graph query",
		"lango graph stats",
		"lango graph clear",
		"lango graph add",
		"lango graph export",
		"lango graph import",
	}
	for _, snippet := range forbiddenAgentMemorySnippets {
		if strings.Contains(agentMemorySection, snippet) {
			t.Fatalf("%s still contains graph command %q inside the Agent & Memory section", target, snippet)
		}
	}
}

func cliIndexGraphDocsScopeGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
