package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIIndexIncludesCoreAndStatusSections(t *testing.T) {
	t.Parallel()

	repoRoot := cliIndexCoreStatusDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "index.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"### Core Commands",
		"lango cockpit",
		"lango serve",
		"lango version",
		"lango health",
		"lango onboard",
		"lango settings",
		"lango doctor",
		"### Status Dashboard",
		"lango status dead-letter-summary",
		"lango status dead-letters",
		"lango status dead-letter <transaction-receipt-id>",
		"lango status dead-letter retry <transaction-receipt-id>",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required CLI index snippet %q", target, snippet)
		}
	}
}

func TestCoreDocsDescribeTUIRuntimeMCPStatus(t *testing.T) {
	t.Parallel()

	repoRoot := cliIndexCoreStatusDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "core.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"configured MCP servers may still initialize through the local interactive bootstrap path",
		"The `/status` slash command reflects that distinction",
		"MCP is shown as active when the local interactive bootstrap initialized MCP",
		"configured-only MCP is shown separately from active runtime features",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required TUI MCP status doc snippet %q", target, snippet)
		}
	}
}

func cliIndexCoreStatusDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
