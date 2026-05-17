package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCmdEntrypointsAvoidDirectOsExitOutsideSeams(t *testing.T) {
	t.Parallel()

	cmdDir := filepath.Join(cmdEntrypointExitGuardRepoRoot(t), "cmd")
	allowedExitLines := map[string][]string{
		filepath.Join("cmd", "lango", "main.go"): {
			"exitFn                             = os.Exit",
		},
		filepath.Join("cmd", "zkexport", "main.go"): {
			"zkexportExitFn           = os.Exit",
		},
	}

	err := filepath.WalkDir(cmdDir, func(path string, d os.DirEntry, walkErr error) error {
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

		relPath, err := filepath.Rel(cmdEntrypointExitGuardRepoRoot(t), path)
		if err != nil {
			return err
		}
		allowed := allowedExitLines[filepath.Clean(relPath)]
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.Contains(line, "os.Exit") {
				continue
			}
			ok := false
			for _, allowedLine := range allowed {
				if strings.Contains(line, allowedLine) {
					ok = true
					break
				}
			}
			if !ok {
				t.Fatalf("%s contains forbidden direct os.Exit reference: %s", path, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan cmd entrypoints for os.Exit regressions: %v", err)
	}
}

func TestInternalCLIPackagesAvoidDirectOsExit(t *testing.T) {
	t.Parallel()

	cliDir := filepath.Join(cmdEntrypointExitGuardRepoRoot(t), "internal", "cli")
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
		for lineNo, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "os.Exit") {
				t.Fatalf("%s:%d contains forbidden direct os.Exit reference: %s", path, lineNo+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan internal CLI packages for os.Exit regressions: %v", err)
	}
}

func cmdEntrypointExitGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
