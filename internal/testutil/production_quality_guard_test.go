package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProductionGoFilesHaveNoContextTODO(t *testing.T) {
	t.Parallel()

	repoRoot := productionQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "cmd"),
		filepath.Join(repoRoot, "internal"),
	}
	needle := "context." + "TODO("

	for _, target := range targets {
		err := filepath.WalkDir(target, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if strings.Contains(path, string(filepath.Separator)+"internal"+string(filepath.Separator)+"testutil"+string(filepath.Separator)) {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), needle) {
				t.Fatalf("%s contains forbidden %q", path, needle)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s for %q: %v", target, needle, err)
		}
	}
}

func TestProductionSchemaCreateCallsStaySerializedAndScoped(t *testing.T) {
	t.Parallel()

	repoRoot := productionQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "cmd"),
		filepath.Join(repoRoot, "internal"),
	}
	needle := "Schema.Create("

	allowed := map[string]bool{
		filepath.Join(repoRoot, "internal", "dbopen", "dbopen.go"):     true,
		filepath.Join(repoRoot, "internal", "session", "ent_store.go"): true,
	}

	var found []string
	for _, target := range targets {
		err := filepath.WalkDir(target, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if strings.Contains(path, string(filepath.Separator)+"internal"+string(filepath.Separator)+"testutil"+string(filepath.Separator)) {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !strings.Contains(string(data), needle) {
				return nil
			}
			found = append(found, path)
			if !allowed[path] {
				t.Fatalf("%s contains forbidden %q outside approved constructors", path, needle)
			}
			if !strings.Contains(string(data), "schemaCreateMu.Lock()") || !strings.Contains(string(data), "schemaCreateMu.Unlock()") {
				t.Fatalf("%s contains %q without schemaCreateMu serialization", path, needle)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s for %q: %v", target, needle, err)
		}
	}

	if len(found) != len(allowed) {
		t.Fatalf("expected exactly %d serialized Schema.Create production call sites, found %d: %v", len(allowed), len(found), found)
	}
}

func productionQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
