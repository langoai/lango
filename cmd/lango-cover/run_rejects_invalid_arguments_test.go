package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-not-a-real-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("stderr = %q, want invalid flag message", stderr.String())
	}
}

func TestRunDiscoversModuleFromGoMod(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	runRejectsInvalidArgumentsWriteFile(t, root, "go.mod", "module example.com/runRejectsInvalidArguments5\n\ngo 1.23\n")
	runRejectsInvalidArgumentsWriteFile(t, root, "pkg/covered.go", "package pkg\n")
	profile := runRejectsInvalidArgumentsWriteCoverage(t, root, []string{
		"mode: set",
		"example.com/runRejectsInvalidArguments5/pkg/covered.go:1.1,2.1 7 1",
		"example.com/runRejectsInvalidArguments5/pkg/covered.go:3.1,4.1 3 0",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"-profile", profile,
		"-root", root,
		"-top", "1",
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Non-generated coverage: 70.00%",
		"Covered statements: 7",
		"Total statements: 10",
		"Uncovered statements: 3",
		"pkg/covered.go",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestRunReturnsErrorForMissingProfile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	runRejectsInvalidArgumentsWriteFile(t, root, "go.mod", "module example.com/runRejectsInvalidArguments5\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"-profile", filepath.Join(root, "missing.out"),
		"-root", root,
	}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{"coverage report failed", "open coverage profile"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr missing %q: %q", want, stderr.String())
		}
	}
}

func TestRunAppliesThresholdGateBranches(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	runRejectsInvalidArgumentsWriteFile(t, root, "go.mod", "module example.com/runRejectsInvalidArguments5\n")
	runRejectsInvalidArgumentsWriteFile(t, root, "pkg/threshold.go", "package pkg\n")
	profile := runRejectsInvalidArgumentsWriteCoverage(t, root, []string{
		"mode: set",
		"example.com/runRejectsInvalidArguments5/pkg/threshold.go:1.1,2.1 90 1",
		"example.com/runRejectsInvalidArguments5/pkg/threshold.go:3.1,4.1 10 0",
	})

	var passingStdout bytes.Buffer
	var passingStderr bytes.Buffer
	passingCode := run([]string{
		"-profile", profile,
		"-root", root,
		"-threshold", "90",
	}, &passingStdout, &passingStderr)

	if passingCode != 0 {
		t.Fatalf("passing exit code = %d, want 0; stderr=%q", passingCode, passingStderr.String())
	}
	if !strings.Contains(passingStdout.String(), "Threshold: 90.00%") {
		t.Fatalf("passing stdout = %q, want threshold summary", passingStdout.String())
	}
	if passingStderr.Len() != 0 {
		t.Fatalf("passing stderr = %q, want empty", passingStderr.String())
	}

	var failingStdout bytes.Buffer
	var failingStderr bytes.Buffer
	failingCode := run([]string{
		"-profile", profile,
		"-root", root,
		"-threshold", "90.01",
	}, &failingStdout, &failingStderr)

	if failingCode != 1 {
		t.Fatalf("failing exit code = %d, want 1", failingCode)
	}
	if !strings.Contains(failingStdout.String(), "Threshold: 90.01%") {
		t.Fatalf("failing stdout = %q, want threshold summary", failingStdout.String())
	}
	for _, want := range []string{"90.00%", "90.01%"} {
		if !strings.Contains(failingStderr.String(), want) {
			t.Fatalf("failing stderr missing %q: %q", want, failingStderr.String())
		}
	}
}

func TestReadModulePathHandlesMissingAndMalformedGoMod(t *testing.T) {
	t.Parallel()

	missingRoot := t.TempDir()
	if got := readModulePath(missingRoot); got != "" {
		t.Fatalf("missing go.mod module = %q, want empty", got)
	}

	malformedRoot := t.TempDir()
	runRejectsInvalidArgumentsWriteFile(t, malformedRoot, "go.mod", "go 1.23\nrequire example.com/dep v1.0.0\n")
	if got := readModulePath(malformedRoot); got != "" {
		t.Fatalf("malformed go.mod module = %q, want empty", got)
	}
}

func runRejectsInvalidArgumentsWriteFile(t *testing.T, root, rel, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
}

func runRejectsInvalidArgumentsWriteCoverage(t *testing.T, root string, lines []string) string {
	t.Helper()

	path := filepath.Join(root, "runRejectsInvalidArguments5.cover.out")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
	return path
}
