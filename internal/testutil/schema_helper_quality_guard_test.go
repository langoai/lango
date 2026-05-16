package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestTestsUseSharedSchemaHelperInsteadOfDirectSchemaCreate(t *testing.T) {
	t.Parallel()

	repoRoot := schemaHelperRepoRoot(t)
	needle := "Schema.Create("
	allowed := map[string]bool{
		filepath.Join(repoRoot, "internal", "testutil", "schemautil", "schemautil.go"):         true,
		filepath.Join(repoRoot, "internal", "testutil", "production_quality_guard_test.go"):    true,
		filepath.Join(repoRoot, "internal", "testutil", "schema_helper_quality_guard_test.go"): true,
	}

	internalRoot := filepath.Join(repoRoot, "internal")
	var violations []string
	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") && path != filepath.Join(repoRoot, "internal", "testutil", "schemautil", "schemautil.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), needle) && !allowed[path] {
			violations = append(violations, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan for direct test Schema.Create usage: %v", err)
	}

	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("direct Schema.Create usage in tests must go through internal/testutil/schemautil: %v", violations)
	}
}

func schemaHelperRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
