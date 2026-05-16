package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedAutomationCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeAutomationDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango cron add",
		"lango cron list",
		"lango cron delete <id-or-name>",
		"lango cron pause <id-or-name>",
		"lango cron resume <id-or-name>",
		"lango cron history",
		"lango workflow run <file>",
		"lango workflow list",
		"lango workflow status <run-id>",
		"lango workflow cancel <run-id>",
		"lango workflow history",
		"lango workflow validate <file>",
		"lango bg list",
		"lango bg status <id>",
		"lango bg cancel <id>",
		"lango bg result <id>",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README automation snippet %q", target, snippet)
		}
	}
}

func readmeAutomationDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
