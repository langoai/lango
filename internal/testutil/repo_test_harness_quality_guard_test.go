package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryTestsAvoidLegacyExecHelpersAndGlobalSTDIO(t *testing.T) {
	t.Parallel()

	repoRoot := repoHarnessGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "cmd"),
		filepath.Join(repoRoot, "internal"),
	}
	forbidden := []string{
		"os.Stdout =",
		"os.Stderr =",
		"os.Stdin =",
		"testutil.ExecCmd(",
		"testutil.ExecCmdOK(",
	}

	err := walkRepositoryTestFiles(targets, func(path string, text string) {
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains forbidden repository test harness pattern %q", path, needle)
			}
		}
	})
	if err != nil {
		t.Fatalf("scan repository tests for forbidden harness patterns: %v", err)
	}
}

func walkRepositoryTestFiles(targets []string, visit func(path string, text string)) error {
	for _, target := range targets {
		err := filepath.WalkDir(target, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			base := filepath.Base(path)
			if base == "cli_test_harness_quality_guard_test.go" || base == "repo_test_harness_quality_guard_test.go" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			visit(path, string(data))
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func repoHarnessGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
