package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGatewayBackedCLIDocsDescribeConfiguredAddressDefaults(t *testing.T) {
	t.Parallel()

	repoRoot := gatewayCLIDocsGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "metrics.md"),
		filepath.Join(repoRoot, "docs", "cli", "alerts.md"),
		filepath.Join(repoRoot, "docs", "cli", "status.md"),
	}
	requiredSnippet := "configured `server.host` and `server.port`"

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		if !strings.Contains(text, requiredSnippet) {
			t.Fatalf("%s must document configured gateway defaults with %q", target, requiredSnippet)
		}
	}
}

func TestGatewayBackedCLIDocsRejectHardcodedOnlyDefaultWording(t *testing.T) {
	t.Parallel()

	repoRoot := gatewayCLIDocsGuardRepoRoot(t)
	forbiddenByFile := map[string][]string{
		filepath.Join(repoRoot, "docs", "cli", "metrics.md"): {
			"| `--addr` | string | `http://localhost:18789` | Gateway address |",
		},
		filepath.Join(repoRoot, "docs", "cli", "alerts.md"): {
			"lango alerts list [--days=7] [--output table|json] [--addr http://localhost:18789]",
			"lango alerts summary [--output table|json] [--addr http://localhost:18789]",
			"| `--addr` | `http://localhost:18789` | Gateway address |",
		},
		filepath.Join(repoRoot, "docs", "cli", "status.md"): {
			"| `--addr` | `http://localhost:18789` | Gateway address to probe for live status |",
		},
	}

	for target, forbiddenSnippets := range forbiddenByFile {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, snippet := range forbiddenSnippets {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s contains stale hardcoded gateway default wording %q", target, snippet)
			}
		}
	}
}

func TestStatusDocsDescribeExplicitAddrProbeAndDisplayTarget(t *testing.T) {
	t.Parallel()

	repoRoot := gatewayCLIDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "status.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)
	requiredSnippets := []string{
		"normalized explicit address",
		"`gateway` field reports that same normalized address",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s must document explicit --addr behavior with %q", target, snippet)
		}
	}
}

func gatewayCLIDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
