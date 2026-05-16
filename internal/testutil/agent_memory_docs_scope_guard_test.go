package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAgentMemoryDocsDelegateGraphCommandsToDedicatedReference(t *testing.T) {
	t.Parallel()

	repoRoot := agentMemoryDocsScopeGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "agent-memory.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"Graph CLI Reference](graph.md)",
		"lango graph status/query/stats/clear/add/export/import",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required scope-handoff snippet %q", target, snippet)
		}
	}

	forbiddenSnippets := []string{
		"### lango graph status",
		"### lango graph query",
		"### lango graph stats",
		"### lango graph clear",
		"### lango graph add",
		"### lango graph export",
		"### lango graph import",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(text, snippet) {
			t.Fatalf("%s still contains duplicated graph command section %q", target, snippet)
		}
	}
}

func agentMemoryDocsScopeGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
