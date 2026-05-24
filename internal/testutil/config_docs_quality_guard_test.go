package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
)

func TestPublicDocsUseCurrentConfigCLIExamples(t *testing.T) {
	t.Parallel()

	repoRoot := configDocsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
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

func TestPublicConfigurationDocsUseDefaultP2PValues(t *testing.T) {
	t.Parallel()

	repoRoot := configDocsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "configuration.md"),
	}
	defaults := config.DefaultConfig()

	requiredRows := []struct {
		key   string
		typ   string
		value any
	}{
		{key: "p2p.listenAddrs", typ: "[]string", value: defaults.P2P.ListenAddrs},
		{key: "p2p.enableRelay", typ: "bool", value: defaults.P2P.EnableRelay},
		{key: "p2p.enableMdns", typ: "bool", value: defaults.P2P.EnableMDNS},
		{key: "p2p.maxPeers", typ: "int", value: defaults.P2P.MaxPeers},
		{key: "p2p.sessionTokenTtl", typ: "duration", value: defaults.P2P.SessionTokenTTL},
		{key: "p2p.zkHandshake", typ: "bool", value: defaults.P2P.ZKHandshake},
		{key: "p2p.zkAttestation", typ: "bool", value: defaults.P2P.ZKAttestation},
		{key: "p2p.toolIsolation.timeoutPerTool", typ: "duration", value: defaults.P2P.ToolIsolation.TimeoutPerTool},
		{key: "p2p.toolIsolation.maxMemoryMB", typ: "int", value: defaults.P2P.ToolIsolation.MaxMemoryMB},
		{key: "p2p.toolIsolation.container.enabled", typ: "bool", value: defaults.P2P.ToolIsolation.Container.Enabled},
		{key: "p2p.toolIsolation.container.runtime", typ: "string", value: defaults.P2P.ToolIsolation.Container.Runtime},
		{key: "p2p.toolIsolation.container.image", typ: "string", value: defaults.P2P.ToolIsolation.Container.Image},
		{key: "p2p.toolIsolation.container.networkMode", typ: "string", value: defaults.P2P.ToolIsolation.Container.NetworkMode},
		{key: "p2p.toolIsolation.container.poolSize", typ: "int", value: defaults.P2P.ToolIsolation.Container.PoolSize},
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		content := string(data)
		for _, row := range requiredRows {
			assertMarkdownConfigRow(t, target, content, row.key, row.typ, []string{markdownDefaultValue(t, row.value)})
		}
	}

	configDocsPath := filepath.Join(repoRoot, "docs", "configuration.md")
	configDocsData, err := os.ReadFile(configDocsPath)
	if err != nil {
		t.Fatalf("read %s: %v", configDocsPath, err)
	}
	configDocsContent := string(configDocsData)
	assertMarkdownConfigRow(t, configDocsPath, configDocsContent, "p2p.zkp.proofCacheDir", "string", []string{defaults.P2P.ZKP.ProofCacheDir})
	assertP2PJSONExampleUsesDefaultValues(t, configDocsPath, configDocsContent, defaults)
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

func markdownDefaultValue(t *testing.T, value any) string {
	t.Helper()

	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case time.Duration:
		return v.String()
	case []string:
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal default value %v: %v", v, err)
		}
		return string(data)
	default:
		t.Fatalf("unsupported markdown default value type %T", value)
		return ""
	}
}

func assertP2PJSONExampleUsesDefaultValues(t *testing.T, path, content string, defaults *config.Config) {
	t.Helper()

	block := markdownJSONBlockAfterHeading(t, content, "## P2P Network")
	var example struct {
		P2P struct {
			ListenAddrs     []string `json:"listenAddrs"`
			EnableRelay     bool     `json:"enableRelay"`
			SessionTokenTTL string   `json:"sessionTokenTtl"`
			ZKHandshake     bool     `json:"zkHandshake"`
			ZKAttestation   bool     `json:"zkAttestation"`
			ZKP             struct {
				ProofCacheDir string `json:"proofCacheDir"`
			} `json:"zkp"`
			ToolIsolation struct {
				MaxMemoryMB int `json:"maxMemoryMB"`
			} `json:"toolIsolation"`
		} `json:"p2p"`
	}
	if err := json.Unmarshal([]byte(block), &example); err != nil {
		t.Fatalf("%s has invalid P2P JSON example: %v", path, err)
	}
	if got, want := example.P2P.ListenAddrs, defaults.P2P.ListenAddrs; strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("%s P2P JSON example listenAddrs = %v, want %v", path, got, want)
	}
	if got, want := example.P2P.EnableRelay, defaults.P2P.EnableRelay; got != want {
		t.Fatalf("%s P2P JSON example enableRelay = %v, want %v", path, got, want)
	}
	if got, want := example.P2P.SessionTokenTTL, defaults.P2P.SessionTokenTTL.String(); got != want {
		t.Fatalf("%s P2P JSON example sessionTokenTtl = %q, want %q", path, got, want)
	}
	if got, want := example.P2P.ZKHandshake, defaults.P2P.ZKHandshake; got != want {
		t.Fatalf("%s P2P JSON example zkHandshake = %v, want %v", path, got, want)
	}
	if got, want := example.P2P.ZKAttestation, defaults.P2P.ZKAttestation; got != want {
		t.Fatalf("%s P2P JSON example zkAttestation = %v, want %v", path, got, want)
	}
	if got, want := example.P2P.ZKP.ProofCacheDir, defaults.P2P.ZKP.ProofCacheDir; got != want {
		t.Fatalf("%s P2P JSON example zkp.proofCacheDir = %q, want %q", path, got, want)
	}
	if got, want := example.P2P.ToolIsolation.MaxMemoryMB, defaults.P2P.ToolIsolation.MaxMemoryMB; got != want {
		t.Fatalf("%s P2P JSON example toolIsolation.maxMemoryMB = %d, want %d", path, got, want)
	}
}

func markdownJSONBlockAfterHeading(t *testing.T, content, heading string) string {
	t.Helper()

	headingIndex := strings.Index(content, heading)
	if headingIndex < 0 {
		t.Fatalf("missing heading %q", heading)
	}
	afterHeading := content[headingIndex+len(heading):]
	blockStart := strings.Index(afterHeading, "```json")
	if blockStart < 0 {
		t.Fatalf("missing JSON code block after heading %q", heading)
	}
	blockContentStart := blockStart + len("```json")
	blockEnd := strings.Index(afterHeading[blockContentStart:], "```")
	if blockEnd < 0 {
		t.Fatalf("missing closing fence for JSON code block after heading %q", heading)
	}
	return strings.TrimSpace(afterHeading[blockContentStart : blockContentStart+blockEnd])
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
