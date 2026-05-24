package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCoreProductionCodeAvoidsUnapprovedDirectStdStreams(t *testing.T) {
	t.Parallel()

	repoRoot := coreStreamGuardRepoRoot(t)
	internalDir := filepath.Join(repoRoot, "internal")
	allowedStdStreamLines := map[string][]string{
		filepath.Join("internal", "approval", "tty.go"): {
			"ttyIsTerminal           = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }",
			"ttyInput      io.Reader = os.Stdin",
			"ttyError      io.Writer = os.Stderr",
		},
		filepath.Join("internal", "bootstrap", "phases.go"): {
			"bootstrapConfirmInput   io.Reader = os.Stdin",
			"bootstrapConfirmOutput  io.Writer = os.Stdout",
			"bootstrapErrWriter      io.Writer = os.Stderr",
		},
		filepath.Join("internal", "logging", "logger.go"): {
			"defaultLogWriter io.Writer = os.Stderr",
		},
		filepath.Join("internal", "observability", "tracing.go"): {
			"var tracingStdoutWriter io.Writer = os.Stderr",
		},
		filepath.Join("internal", "sandbox", "worker.go"): {
			"workerStdin  io.Reader = os.Stdin",
			"workerStdout io.Writer = os.Stdout",
		},
		filepath.Join("internal", "security", "passphrase", "acquire.go"): {
			"passphraseStdin         io.Reader = os.Stdin",
			"passphraseStderr        io.Writer = os.Stderr",
		},
		filepath.Join("internal", "security", "passphrase", "stdin.go"): {
			"return ReadStdinPipeFromReader(os.Stdin)",
		},
		filepath.Join("internal", "storagebroker", "client.go"): {
			"var brokerStderr io.Writer = os.Stderr",
		},
		filepath.Join("internal", "tools", "exec", "exec.go"): {
			"var execWarningWriter io.Writer = os.Stderr",
		},
	}

	if err := scanCoreProductionStreams(repoRoot, internalDir, allowedStdStreamLines); err != nil {
		t.Fatalf("scan core production code for stream regressions: %v", err)
	}
}

func TestCoreProductionStreamGuardRejectsDirectStdout(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	coreDir := filepath.Join(root, "internal", "bad")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	target := filepath.Join(coreDir, "bad.go")
	source := []byte(`package bad

import "os"

var output = os.Stdout
`)
	if err := os.WriteFile(target, source, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	err := scanCoreProductionStreams(root, filepath.Join(root, "internal"), nil)
	if err == nil {
		t.Fatal("expected direct os.Stdout reference to be rejected")
	}
	if !strings.Contains(err.Error(), "os.Stdout") {
		t.Fatalf("expected error to identify direct stdout reference, got %v", err)
	}
}

func scanCoreProductionStreams(repoRoot, internalDir string, allowedStdStreamLines map[string][]string) error {
	return filepath.WalkDir(internalDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if shouldSkipCoreStreamDir(repoRoot, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return scanCoreProductionStreamFile(repoRoot, allowedStdStreamLines, path)
	})
}

func scanCoreProductionStreamFile(repoRoot string, allowedStdStreamLines map[string][]string, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return err
	}
	allowed := allowedStdStreamLines[filepath.Clean(relPath)]
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		streamRef := findStdStreamRef(line)
		if streamRef == "" {
			continue
		}
		if isAllowedStdStreamLine(line, allowed) {
			continue
		}
		return fmt.Errorf("%s contains unapproved direct %s reference: %s", path, streamRef, trimmed)
	}
	return nil
}

func shouldSkipCoreStreamDir(repoRoot, path string) bool {
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return false
	}
	clean := filepath.Clean(relPath)
	return clean == filepath.Join("internal", "cli") ||
		clean == filepath.Join("internal", "testutil") ||
		clean == filepath.Join("internal", "ent")
}

func findStdStreamRef(line string) string {
	for _, ref := range []string{"os.Stdin", "os.Stdout", "os.Stderr"} {
		if strings.Contains(line, ref) {
			return ref
		}
	}
	return ""
}

func isAllowedStdStreamLine(line string, allowedLines []string) bool {
	for _, allowedLine := range allowedLines {
		if strings.Contains(line, allowedLine) {
			return true
		}
	}
	return false
}

func coreStreamGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
