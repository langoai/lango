package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCmdEntrypointsAvoidRawPrintsAndDirectStdStreams(t *testing.T) {
	t.Parallel()

	cmdDir := filepath.Join(cmdEntrypointGuardRepoRoot(t), "cmd")
	allowedStdStreamLines := map[string][]string{
		filepath.Join("cmd", "lango", "main.go"): {
			"mainStdin                io.Reader = os.Stdin",
			"mainStdout               io.Writer = os.Stdout",
			"mainStderr               io.Writer = os.Stderr",
			"cockpitStartupErrWriter   io.Writer = os.Stderr",
			"workbenchStartupErrWriter io.Writer = os.Stderr",
			"chatStartupErrWriter      io.Writer = os.Stderr",
		},
		filepath.Join("cmd", "zkexport", "main.go"): {
			"zkexportStdout io.Writer = os.Stdout",
			"zkexportStderr io.Writer = os.Stderr",
		},
	}
	forbiddenPrints := []string{
		"fmt.Print(",
		"fmt.Printf(",
		"fmt.Println(",
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
		text := string(data)
		for _, needle := range forbiddenPrints {
			if strings.Contains(text, needle) {
				t.Fatalf("%s contains forbidden raw print call %q", path, needle)
			}
		}

		relPath, err := filepath.Rel(cmdEntrypointGuardRepoRoot(t), path)
		if err != nil {
			return err
		}
		allowed := allowedStdStreamLines[filepath.Clean(relPath)]
		for _, line := range strings.Split(text, "\n") {
			if !(strings.Contains(line, "os.Stdout") || strings.Contains(line, "os.Stderr") || strings.Contains(line, "os.Stdin")) {
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
				t.Fatalf("%s contains forbidden direct standard-stream reference: %s", path, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan cmd entrypoints for stream regressions: %v", err)
	}
}

func cmdEntrypointGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
