package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestContractCLIDocsUseCurrentOutputFormatContract(t *testing.T) {
	t.Parallel()

	repoRoot := contractCLIDocsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "contract.md"),
		filepath.Join(repoRoot, "openspec", "specs", "contract-interaction", "spec.md"),
	}

	forbiddenPatterns := []struct {
		re     *regexp.Regexp
		reason string
	}{
		{
			re:     regexp.MustCompile(`\| \x60--output\x60 \| bool \| \x60false\x60 \| Output as JSON \|`),
			reason: "stale boolean --output flag table entry",
		},
		{
			re:     regexp.MustCompile(`(?m)--output\s*$`),
			reason: "stale bare --output example without explicit format",
		},
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		for _, pattern := range forbiddenPatterns {
			if pattern.re.Find(data) != nil {
				t.Fatalf("%s contains %s", target, pattern.reason)
			}
		}
	}
}

func contractCLIDocsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
