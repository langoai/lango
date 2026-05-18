package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_PrintsReport(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "internal/a.go")
	writeSource(t, root, "internal/b.go")
	profile := writeCoverage(t, root, []string{
		"mode: set",
		"github.com/langoai/lango/internal/a.go:1.1,2.1 10 1",
		"github.com/langoai/lango/internal/a.go:3.1,4.1 5 0",
		"github.com/langoai/lango/internal/b.go:1.1,2.1 20 0",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"-profile", profile,
		"-root", root,
		"-module", "github.com/langoai/lango",
		"-top", "2",
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Non-generated coverage: 28.57%",
		"Covered statements: 10",
		"Total statements: 35",
		"Uncovered statements: 25",
		"internal/b.go",
		"internal/a.go",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestRun_ReturnsNonZeroWhenThresholdFails(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "internal/a.go")
	profile := writeCoverage(t, root, []string{
		"mode: set",
		"github.com/langoai/lango/internal/a.go:1.1,2.1 89 1",
		"github.com/langoai/lango/internal/a.go:3.1,4.1 11 0",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"-profile", profile,
		"-root", root,
		"-module", "github.com/langoai/lango",
		"-threshold", "90",
	}, &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero; stdout=%q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "89.00%") || !strings.Contains(stderr.String(), "90.00%") {
		t.Fatalf("stderr = %q, want measured and required percentages", stderr.String())
	}
}

func TestRun_ReturnsZeroWhenThresholdPasses(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "internal/a.go")
	profile := writeCoverage(t, root, []string{
		"mode: set",
		"github.com/langoai/lango/internal/a.go:1.1,2.1 90 1",
		"github.com/langoai/lango/internal/a.go:3.1,4.1 10 0",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"-profile", profile,
		"-root", root,
		"-module", "github.com/langoai/lango",
		"-threshold", "90",
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Threshold: 90.00%") {
		t.Fatalf("stdout = %q, want threshold summary", stdout.String())
	}
}

func writeSource(t *testing.T, root, rel string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("package internal\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
}

func writeCoverage(t *testing.T, root string, lines []string) string {
	t.Helper()

	path := filepath.Join(root, "coverage.out")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
	return path
}
