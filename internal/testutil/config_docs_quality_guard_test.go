package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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

func TestPublicConfigurationDocsIncludeCloudKMSSettings(t *testing.T) {
	t.Parallel()

	repoRoot := configDocsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "configuration.md"),
	}

	requiredRows := []struct {
		key      string
		typ      string
		defaults []string
	}{
		{key: "security.kms.region", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.keyId", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.endpoint", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.fallbackToLocal", typ: "bool", defaults: []string{"`true`"}},
		{key: "security.kms.timeoutPerOperation", typ: "duration", defaults: []string{"`5s`"}},
		{key: "security.kms.maxRetries", typ: "int", defaults: []string{"`3`"}},
		{key: "security.kms.azure.vaultUrl", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.azure.keyVersion", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.pkcs11.modulePath", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.pkcs11.slotId", typ: "int", defaults: []string{"`0`"}},
		{key: "security.kms.pkcs11.pin", typ: "string", defaults: []string{"", "-"}},
		{key: "security.kms.pkcs11.keyLabel", typ: "string", defaults: []string{"", "-"}},
	}
	requiredNotes := []string{
		"`LANGO_KMS_PROVIDER`",
		"`LANGO_KMS_FALLBACK_TO_LOCAL=false`",
		"before profile config is loaded",
	}
	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		content := string(data)
		for _, row := range requiredRows {
			assertMarkdownConfigRow(t, target, content, row.key, row.typ, row.defaults)
		}
		for _, want := range requiredNotes {
			if !strings.Contains(content, want) {
				t.Fatalf("%s missing Cloud KMS configuration docs token %q", target, want)
			}
		}
	}
}

func assertMarkdownConfigRow(t *testing.T, path, content, key, wantType string, wantDefaults []string) {
	t.Helper()

	needle := "`" + key + "`"
	for _, line := range strings.Split(content, "\n") {
		cols := splitMarkdownRow(line)
		if len(cols) < 4 || cols[0] != needle {
			continue
		}
		if normalizeMarkdownCell(cols[1]) != wantType {
			t.Fatalf("%s row %s has type %q, want %q", path, key, cols[1], wantType)
		}
		if !containsMarkdownCell(wantDefaults, cols[2]) {
			t.Fatalf("%s row %s has default %q, want one of %v", path, key, cols[2], wantDefaults)
		}
		if cols[3] == "" {
			t.Fatalf("%s row %s has empty description", path, key)
		}
		return
	}
	t.Fatalf("%s missing Cloud KMS configuration docs row %q", path, needle)
}

func splitMarkdownRow(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
		return nil
	}
	parts := strings.Split(strings.Trim(line, "|"), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func containsMarkdownCell(values []string, want string) bool {
	normalizedWant := normalizeMarkdownCell(want)
	for _, value := range values {
		if normalizeMarkdownCell(value) == normalizedWant {
			return true
		}
	}
	return false
}

func normalizeMarkdownCell(value string) string {
	return strings.Trim(strings.TrimSpace(value), "`")
}

func configDocsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
