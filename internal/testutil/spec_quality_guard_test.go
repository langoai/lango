package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMainSpecsHaveNoArchivedPurposePlaceholders(t *testing.T) {
	t.Parallel()

	specsDir := filepath.Join(specQualityRepoRoot(t), "openspec", "specs")
	forbidden := []string{
		"TBD - created by archiving change",
		"Update Purpose after archive.",
	}

	err := filepath.WalkDir(specsDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Base(path) != "spec.md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains forbidden archived-purpose placeholder %q", path, needle)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan specs for archived-purpose placeholders: %v", err)
	}
}

func specQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
