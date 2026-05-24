package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExtensionCLIDocsDescribeCurrentCommandSurface(t *testing.T) {
	t.Parallel()

	repoRoot := extensionCLIDocsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "extension.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango extension inspect <source>",
		"lango extension install <source>",
		"lango extension list",
		"lango extension remove <name>",
		"table|json|plain",
		"pass `--yes`",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required extension docs snippet %q", target, snippet)
		}
	}
}

func extensionCLIDocsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
