package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIProductionCodeAvoidsDuplicatePrettyJSONWriters(t *testing.T) {
	t.Parallel()

	repoRoot := cliPrettyJSONGuardRepoRoot(t)
	cliDir := filepath.Join(repoRoot, "internal", "cli")
	allowed := map[string]struct{}{
		filepath.Join("internal", "cli", "clihttp", "clihttp.go"): {},
	}
	needle := `SetIndent("", "  ")`

	err := filepath.WalkDir(cliDir, func(path string, d os.DirEntry, walkErr error) error {
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
		if !strings.Contains(string(data), needle) {
			return nil
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if _, ok := allowed[filepath.Clean(relPath)]; ok {
			return nil
		}

		t.Fatalf("%s contains duplicate pretty-JSON writer setup %q", path, needle)
		return nil
	})
	if err != nil {
		t.Fatalf("scan CLI production code for duplicate pretty-JSON writers: %v", err)
	}
}

func cliPrettyJSONGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
