package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPackageConsolidationSpecAvoidsDeletedPackagePathClaims(t *testing.T) {
	t.Parallel()

	repoRoot := packageConsolidationSpecGuardRepoRoot(t)
	specPath := filepath.Join(repoRoot, "openspec", "specs", "package-consolidation", "spec.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	text := string(data)

	forbidden := []string{
		"`internal/ctxutil/`",
		"`internal/passphrase/`",
		"`internal/zkp/`",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			t.Fatalf("%s contains stale deleted package path %q", specPath, needle)
		}
	}
}

func packageConsolidationSpecGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
