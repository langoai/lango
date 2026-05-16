package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIProductionCodeAvoidsBooleanOutputFlags(t *testing.T) {
	t.Parallel()

	repoRoot := cliOutputFlagGuardRepoRoot(t)
	cliDir := filepath.Join(repoRoot, "internal", "cli")
	forbidden := []string{
		`BoolVar(&`,
		`BoolVarP(&`,
	}

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
		text := string(data)
		for _, prefix := range forbidden {
			if strings.Contains(text, prefix+`jsonOutput, "output"`) || strings.Contains(text, prefix+`asJSON, "output"`) || strings.Contains(text, prefix+`output, "output"`) {
				t.Fatalf("%s contains forbidden boolean --output flag declaration", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan CLI production code for boolean --output flags: %v", err)
	}
}

func cliOutputFlagGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
