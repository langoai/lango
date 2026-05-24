package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestMigratedCLIProductionCodeAvoidsBooleanJSONFlags(t *testing.T) {
	t.Parallel()

	repoRoot := cliJSONFlagMigrationGuardRepoRoot(t)
	targetDirs := []string{
		filepath.Join(repoRoot, "internal", "cli", "a2a"),
		filepath.Join(repoRoot, "internal", "cli", "approval"),
		filepath.Join(repoRoot, "internal", "cli", "contract"),
		filepath.Join(repoRoot, "internal", "cli", "doctor"),
		filepath.Join(repoRoot, "internal", "cli", "graph"),
		filepath.Join(repoRoot, "internal", "cli", "learning"),
		filepath.Join(repoRoot, "internal", "cli", "librarian"),
		filepath.Join(repoRoot, "internal", "cli", "memory"),
		filepath.Join(repoRoot, "internal", "cli", "payment"),
		filepath.Join(repoRoot, "internal", "cli", "run"),
		filepath.Join(repoRoot, "internal", "cli", "security"),
		filepath.Join(repoRoot, "internal", "cli", "workflow"),
	}

	jsonFlagDecl := regexp.MustCompile(`BoolVarP?\([^)\n]*"json"`)

	for _, targetDir := range targetDirs {
		err := filepath.WalkDir(targetDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if jsonFlagDecl.Find(data) != nil {
				t.Fatalf("%s reintroduces a boolean --json flag declaration", path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s for boolean --json flag regressions: %v", targetDir, err)
		}
	}
}

func cliJSONFlagMigrationGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
