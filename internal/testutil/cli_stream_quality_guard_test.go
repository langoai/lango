package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIProductionCodeAvoidsRawPrintsAndDirectStdStreams(t *testing.T) {
	t.Parallel()

	cliDir := filepath.Join(cliStreamGuardRepoRoot(t), "internal", "cli")
	allowedStdStreamRefs := map[string]struct{}{
		filepath.Join("internal", "cli", "prompt", "prompt.go"):          {},
		filepath.Join("internal", "cli", "security", "status.go"):       {},
	}
	forbidden := []string{
		"fmt.Print(",
		"fmt.Printf(",
		"fmt.Println(",
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
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains forbidden raw print call %q", path, needle)
			}
		}

		relPath, err := filepath.Rel(cliStreamGuardRepoRoot(t), path)
		if err != nil {
			return err
		}
		if _, ok := allowedStdStreamRefs[filepath.Clean(relPath)]; ok {
			return nil
		}
		if strings.Contains(text, "os.Stdout") || strings.Contains(text, "os.Stderr") {
			t.Fatalf("%s contains forbidden direct os.Stdout/os.Stderr reference", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan CLI production code for stream regressions: %v", err)
	}
}

func cliStreamGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
