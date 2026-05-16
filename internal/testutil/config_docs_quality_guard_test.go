package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestPublicDocsUseCurrentConfigCLIExamples(t *testing.T) {
	t.Parallel()

	repoRoot := configDocsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs"),
	}

	forbiddenPatterns := []struct {
		re     *regexp.Regexp
		reason string
	}{
		{
			re:     regexp.MustCompile(`lango config get\s+[^\\\n` + "`" + `]+--format json`),
			reason: "stale config get --format json example",
		},
		{
			re:     regexp.MustCompile(`lango config export\s*>`),
			reason: "profile-less config export example",
		},
		{
			re:     regexp.MustCompile(`(?m)^lango config import\s+\S+(?:\s*)$`),
			reason: "profile-less config import example",
		},
	}

	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("stat %s: %v", target, err)
		}
		if info.IsDir() {
			err = filepath.WalkDir(target, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				checkNoForbiddenConfigDocPattern(t, path, forbiddenPatterns)
				return nil
			})
			if err != nil {
				t.Fatalf("walk %s: %v", target, err)
			}
			continue
		}
		checkNoForbiddenConfigDocPattern(t, target, forbiddenPatterns)
	}
}

func checkNoForbiddenConfigDocPattern(t *testing.T, path string, patterns []struct {
	re     *regexp.Regexp
	reason string
}) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, pattern := range patterns {
		if pattern.re.Find(data) != nil {
			t.Fatalf("%s contains %s", path, pattern.reason)
		}
	}
}

func configDocsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
