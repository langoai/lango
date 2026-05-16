package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCoreDocsDelegateAgentDiagnosticsToDedicatedReference(t *testing.T) {
	t.Parallel()

	repoRoot := coreDocsScopeGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "core.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"Agent CLI Reference](agent.md)",
		"lango agent trace list",
		"lango agent trace show <trace-id>",
		"lango agent trace metrics",
		"lango agent graph <session-key>",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required scope-handoff snippet %q", target, snippet)
		}
	}

	forbiddenSnippets := []string{
		"## lango agent trace",
		"## lango agent graph",
		"## lango agent trace metrics",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(text, snippet) {
			t.Fatalf("%s still contains duplicated agent diagnostics section %q", target, snippet)
		}
	}
}

func coreDocsScopeGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
