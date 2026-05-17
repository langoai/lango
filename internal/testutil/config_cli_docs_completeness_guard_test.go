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
		"--show-secrets",
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

	configDoc := filepath.Join(repoRoot, "docs", "cli", "config.md")
	data, err := os.ReadFile(configDoc)
	if err != nil {
		t.Fatalf("read %s: %v", configDoc, err)
	}
	text := string(data)
	configSnippets := []string{
		"lango config get <dot.path> [--output plain|json] [--show-secrets]",
		"Sensitive scalar paths and nested sensitive fields inside object or map reads are redacted by default.",
	}
	for _, snippet := range configSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required config get security doc snippet %q", configDoc, snippet)
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
