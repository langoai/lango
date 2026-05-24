package testutil

import (
	"fmt"
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
		filepath.Join("internal", "cli", "prompt", "prompt.go"):   {},
		filepath.Join("internal", "cli", "security", "status.go"): {},
	}
	err := filepath.WalkDir(cliDir, func(path string, d os.DirEntry, walkErr error) error {
		return scanCLIProductionStreamFile(cliStreamGuardRepoRoot(t), allowedStdStreamRefs, path, d, walkErr)
	})
	if err != nil {
		t.Fatalf("scan CLI production code for stream regressions: %v", err)
	}
}

func scanCLIProductionStreams(repoRoot, cliDir string, allowedStdStreamRefs map[string]struct{}) error {
	return filepath.WalkDir(cliDir, func(path string, d os.DirEntry, walkErr error) error {
		return scanCLIProductionStreamFile(repoRoot, allowedStdStreamRefs, path, d, walkErr)
	})
}

func scanCLIProductionStreamFile(repoRoot string, allowedStdStreamRefs map[string]struct{}, path string, d os.DirEntry, walkErr error) error {
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
	forbidden := []string{
		"fmt.Print(",
		"fmt.Printf(",
		"fmt.Println(",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			return fmt.Errorf("%s contains forbidden raw print call %q", path, needle)
		}
	}

	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return err
	}
	if _, ok := allowedStdStreamRefs[filepath.Clean(relPath)]; ok {
		return nil
	}
	if strings.Contains(text, "os.Stdin") || strings.Contains(text, "os.Stdout") || strings.Contains(text, "os.Stderr") {
		return fmt.Errorf("%s contains forbidden direct standard-stream reference", path)
	}
	return nil
}

func TestCLIProductionStreamGuardRejectsDirectStdin(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cliDir := filepath.Join(root, "internal", "cli", "bad")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	target := filepath.Join(cliDir, "bad.go")
	source := []byte(`package bad

import "os"

var input = os.Stdin
`)
	if err := os.WriteFile(target, source, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if err := scanCLIProductionStreams(root, filepath.Join(root, "internal", "cli"), nil); err == nil {
		t.Fatal("expected direct os.Stdin reference to be rejected")
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
