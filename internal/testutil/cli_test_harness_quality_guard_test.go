package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLITestsAvoidGlobalSTDIOAndLegacyExecHelpers(t *testing.T) {
	t.Parallel()

	cliDir := filepath.Join(cliHarnessGuardRepoRoot(t), "internal", "cli")
	forbidden := []string{
		"os.Stdout =",
		"os.Stderr =",
		"os.Stdin =",
		"testutil.ExecCmd(",
		"testutil.ExecCmdOK(",
	}

	err := filepath.WalkDir(cliDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains forbidden CLI test harness pattern %q", path, needle)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan CLI tests for forbidden harness patterns: %v", err)
	}
}

func cliHarnessGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
