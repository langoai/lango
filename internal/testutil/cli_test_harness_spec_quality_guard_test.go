package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLITestHarnessSpecAvoidsDeletedHarnessPathClaims(t *testing.T) {
	t.Parallel()

	repoRoot := cliHarnessSpecGuardRepoRoot(t)
	specPath := filepath.Join(repoRoot, "openspec", "specs", "cli-test-harness", "spec.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	text := string(data)

	forbidden := []string{
		"internal/testutil/cli_harness.go",
		"stdout/stderr capture, and cobra command execution helper",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			t.Fatalf("%s contains stale CLI harness claim %q", specPath, needle)
		}
	}

	required := []string{
		"internal/testutil/loaders.go",
		"internal/testutil/helpers.go",
		"testutil.FakeBootLoader(t, cfg)",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("%s is missing current CLI harness reference %q", specPath, needle)
		}
	}
}

func cliHarnessSpecGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
