package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedMCPCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmeMCPDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "index.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango mcp list",
		"lango mcp add <name>",
		"lango mcp remove <name>",
		"lango mcp get <name>",
		"lango mcp test <name>",
		"lango mcp enable <name>",
		"lango mcp disable <name>",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README MCP snippet %q", target, snippet)
		}
	}
}

func readmeMCPDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
