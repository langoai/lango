package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPublicConfigCLIDocsIncludeImplementedReadWriteCommands(t *testing.T) {
	t.Parallel()

	repoRoot := configCLIDocsRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "cli", "config.md"),
	}
	requiredSnippets := []string{
		"lango config get <dot.path>",
		"lango config set <dot.path> [value]",
		"--from-env ENV",
		"lango config keys [prefix]",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing required config CLI doc snippet %q", target, snippet)
			}
		}
	}
}

func configCLIDocsRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
