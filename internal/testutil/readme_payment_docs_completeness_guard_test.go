package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEIncludesImplementedPaymentCommands(t *testing.T) {
	t.Parallel()

	repoRoot := readmePaymentDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "index.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango payment balance",
		"lango payment history",
		"lango payment limits",
		"lango payment info",
		"lango payment send",
		"lango payment x402",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required README payment snippet %q", target, snippet)
		}
	}
}

func readmePaymentDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
