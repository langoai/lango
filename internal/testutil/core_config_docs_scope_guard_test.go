package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCoreDocsDelegateConfigCommandsToDedicatedReference(t *testing.T) {
	t.Parallel()

	repoRoot := coreConfigDocsScopeGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "core.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"Config Management](config.md)",
		"lango config list/create/use/delete/import/export/get/set/keys/validate",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing required scope-handoff snippet %q", target, snippet)
		}
	}

	forbiddenSnippets := []string{
		"### lango config list",
		"### lango config create",
		"### lango config use",
		"### lango config delete",
		"### lango config import",
		"### lango config export",
		"### lango config validate",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(text, snippet) {
			t.Fatalf("%s still contains duplicated config command section %q", target, snippet)
		}
	}
}

func coreConfigDocsScopeGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
