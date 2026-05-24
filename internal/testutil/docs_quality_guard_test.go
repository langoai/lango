package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestPublicDocsHaveNoStaleSharedConfirmExamples(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs"),
	}
	staleSharedConfirmExample := regexp.MustCompile(`\[y/N\] (y|n|yes|no)\b`)

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
				checkNoStaleConfirmExample(t, path, staleSharedConfirmExample)
				return nil
			})
			if err != nil {
				t.Fatalf("walk %s: %v", target, err)
			}
			continue
		}
		checkNoStaleConfirmExample(t, target, staleSharedConfirmExample)
	}
}

func TestArchitectureDocsAvoidStaleBrokenCodePaths(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "architecture", "data-flow.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}

	stalePath := []byte("`internal/librarian/buffer.go`")
	if regexp.MustCompile(regexp.QuoteMeta(string(stalePath))).Find(data) != nil {
		t.Fatalf("%s contains stale broken code path %q", target, string(stalePath))
	}
}

func TestArchitectureProjectStructureAvoidsDeletedPassphrasePackagePath(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	if regexp.MustCompile(regexp.QuoteMeta("`passphrase/`")).MatchString(text) {
		t.Fatalf("%s still contains deleted package path `passphrase/`", target)
	}
	if !regexp.MustCompile(regexp.QuoteMeta("`security/passphrase/`")).MatchString(text) {
		t.Fatalf("%s does not describe the current security/passphrase package path", target)
	}
}

func TestArchitectureProjectStructureUsesCurrentSecurityCLISurface(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	if !regexp.MustCompile(regexp.QuoteMeta("`lango security status`, `change-passphrase`, deprecated `migrate-passphrase`")).MatchString(text) {
		t.Fatalf("%s does not describe the current canonical/deprecated security CLI surface", target)
	}
	requiredSurfaceSnippets := []string{
		"`keyring store/clear/status`",
		"`recovery setup/restore`",
		"`kms status/test/keys/wrap/detach`",
	}
	for _, snippet := range requiredSurfaceSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(text) {
			t.Fatalf("%s does not describe current security inventory snippet %q", target, snippet)
		}
	}
}

func TestPublicDocsExplainBackgroundCLIGatewayBoundary(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "automation", "background.md"),
	}
	requiredSnippets := []string{
		"Background task state remains in-memory and owned by the target running app/server process",
		"Root `lango bg` commands talk to that process through the Lango gateway",
		"use `--addr <url>` to target a non-default gateway",
		"otherwise the CLI uses the configured server host/port",
		"Server restart clears tasks",
		"Auth-enabled gateways require gateway session authentication and reject unauthenticated root CLI background requests",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		if !strings.Contains(text, "lango bg list") {
			t.Fatalf("%s no longer lists lango bg commands; update this guard if the public surface changes", target)
		}
		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing background CLI gateway-boundary caveat snippet %q", target, snippet)
			}
		}
	}
}

func TestPublicDocsExplainBackgroundCLIMutability(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "automation", "background.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	staleReadOnlyCLI := regexp.MustCompile(`(?is)\bCLI\b.{0,80}\bread-only\b.{0,80}\bbackground tasks\b`)
	if staleReadOnlyCLI.MatchString(text) {
		t.Fatalf("%s still describes background CLI management as read-only", target)
	}

	requiredPatterns := map[string]*regexp.Regexp{
		"gateway-backed management": regexp.MustCompile(`(?is)CLI\b.{0,80}\bgateway-backed management commands\b`),
		"inspect commands":          regexp.MustCompile(`(?is)lango bg list\b.{0,80}lango bg status <id>.{0,80}lango bg result <id>.{0,80}\binspect task state\b.{0,80}\btarget gateway process\b`),
		"cancel mutability":         regexp.MustCompile(`(?is)lango bg cancel <id>.{0,120}\brequests cancellation\b.{0,120}\bpending or running task\b.{0,120}\btarget gateway process\b`),
		"submission boundary": regexp.MustCompile(
			regexp.QuoteMeta("Task submission is handled exclusively through agent tools"),
		),
		"gateway boundary": regexp.MustCompile(`(?is)Root ` + "`" + `lango bg` + "`" + ` commands\b.{0,120}\bLango gateway\b`),
	}
	for name, pattern := range requiredPatterns {
		if !pattern.MatchString(text) {
			t.Fatalf("%s is missing background CLI mutability contract %q", target, name)
		}
	}
}

func TestPublicDocsExplainBareRootNonInteractiveFallback(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "cli", "core.md"),
	}
	requiredSnippets := []string{
		"Interactive bare `lango` starts the mission workbench TUI",
		"Non-interactive bare `lango` prints help to command stdout and exits successfully without starting the TUI",
		"Unlike `lango cockpit` and `lango chat`, this bare-root fallback is not an actionable non-interactive error",
	}

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		text := string(data)
		for _, snippet := range requiredSnippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s is missing bare-root non-interactive fallback snippet %q", target, snippet)
			}
		}
	}
}

func TestP2PDocsExplainConnectTimeoutContract(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "p2p.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"`lango p2p connect` uses `p2p.handshakeTimeout` as its connection timeout",
		"falls back to 30 seconds when that setting is unset or invalid",
		"Canceling the command context also cancels the in-flight connect attempt",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing P2P connect timeout contract snippet %q", target, snippet)
		}
	}
}

func TestP2PDocsExplainEphemeralStartupCancellationContract(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "p2p.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"Ephemeral-node commands honor the Cobra command context during startup",
		"Canceling the command context cancels DHT bootstrap, bootstrap peer dials, and mDNS discovered-peer connection attempts for that temporary node",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing P2P startup cancellation contract snippet %q", target, snippet)
		}
	}
}

func TestArchitectureProjectStructureUsesCurrentGraphAndMetricsCLISurface(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	graphSnippet := "`lango graph status`, `query`, `stats`, `clear`, `add`, `export`, `import`"
	if !regexp.MustCompile(regexp.QuoteMeta(graphSnippet)).MatchString(text) {
		t.Fatalf("%s does not describe the current graph CLI surface", target)
	}

	metricsSnippet := "`lango metrics`, `sessions`, `tools`, `agents`, `policy`, `history`"
	if !regexp.MustCompile(regexp.QuoteMeta(metricsSnippet)).MatchString(text) {
		t.Fatalf("%s does not describe the current metrics CLI surface", target)
	}
}

func TestArchitectureAndREADMEUseCurrentMemoryInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	if !regexp.MustCompile(regexp.QuoteMeta("`lango memory list`, `status`, `clear`, `agents`, `agent <name>`")).MatchString(architectureText) {
		t.Fatalf("%s does not describe the current memory CLI surface", architectureTarget)
	}
}

func TestArchitectureAndREADMEUseCurrentContractInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	if !regexp.MustCompile(regexp.QuoteMeta("`lango contract read`, `call`, `abi load`")).MatchString(architectureText) {
		t.Fatalf("%s does not describe the current contract CLI surface", architectureTarget)
	}
}

func TestCLIIndexLinksEveryDedicatedCLIReference(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	indexTarget := filepath.Join(repoRoot, "docs", "cli", "index.md")
	indexData, err := os.ReadFile(indexTarget)
	if err != nil {
		t.Fatalf("read %s: %v", indexTarget, err)
	}
	indexText := string(indexData)

	cliDocsDir := filepath.Join(repoRoot, "docs", "cli")
	entries, err := os.ReadDir(cliDocsDir)
	if err != nil {
		t.Fatalf("read %s: %v", cliDocsDir, err)
	}

	var missing []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "index.md" {
			continue
		}
		if !regexp.MustCompile(regexp.QuoteMeta("(" + entry.Name() + ")")).MatchString(indexText) {
			missing = append(missing, entry.Name())
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("%s is missing dedicated CLI reference links for %v", indexTarget, missing)
	}
}

func TestArchitectureIndexLinksEveryArchitectureReference(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	indexTarget := filepath.Join(repoRoot, "docs", "architecture", "index.md")
	indexData, err := os.ReadFile(indexTarget)
	if err != nil {
		t.Fatalf("read %s: %v", indexTarget, err)
	}
	indexText := string(indexData)

	architectureDir := filepath.Join(repoRoot, "docs", "architecture")
	entries, err := os.ReadDir(architectureDir)
	if err != nil {
		t.Fatalf("read %s: %v", architectureDir, err)
	}

	var missing []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "index.md" {
			continue
		}
		if !regexp.MustCompile(regexp.QuoteMeta(entry.Name())).MatchString(indexText) {
			missing = append(missing, entry.Name())
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("%s is missing architecture reference links for %v", indexTarget, missing)
	}
}

func TestDocsHomeLinksEveryTopLevelSectionIndex(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	indexTarget := filepath.Join(repoRoot, "docs", "index.md")
	indexData, err := os.ReadFile(indexTarget)
	if err != nil {
		t.Fatalf("read %s: %v", indexTarget, err)
	}
	indexText := string(indexData)

	docsRoot := filepath.Join(repoRoot, "docs")
	entries, err := os.ReadDir(docsRoot)
	if err != nil {
		t.Fatalf("read %s: %v", docsRoot, err)
	}

	var missing []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		indexName := filepath.Join(entry.Name(), "index.md")
		if _, err := os.Stat(filepath.Join(docsRoot, indexName)); err != nil {
			continue
		}
		if !regexp.MustCompile(regexp.QuoteMeta(indexName)).MatchString(indexText) {
			missing = append(missing, indexName)
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("%s is missing top-level section links for %v", indexTarget, missing)
	}
}

func TestFeaturesIndexLinksEveryFeatureReference(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	indexTarget := filepath.Join(repoRoot, "docs", "features", "index.md")
	indexData, err := os.ReadFile(indexTarget)
	if err != nil {
		t.Fatalf("read %s: %v", indexTarget, err)
	}
	indexText := string(indexData)

	featuresDir := filepath.Join(repoRoot, "docs", "features")
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		t.Fatalf("read %s: %v", featuresDir, err)
	}

	var missing []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "index.md" {
			continue
		}
		if !regexp.MustCompile(regexp.QuoteMeta(entry.Name())).MatchString(indexText) {
			missing = append(missing, entry.Name())
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("%s is missing feature reference links for %v", indexTarget, missing)
	}
}

func TestEveryDocsSectionIndexLinksItsOwnPages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	docsRoot := filepath.Join(repoRoot, "docs")

	entries, err := os.ReadDir(docsRoot)
	if err != nil {
		t.Fatalf("read %s: %v", docsRoot, err)
	}

	var failures []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sectionDir := filepath.Join(docsRoot, entry.Name())
		indexPath := filepath.Join(sectionDir, "index.md")
		if _, err := os.Stat(indexPath); err != nil {
			continue
		}

		indexData, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("read %s: %v", indexPath, err)
		}
		indexText := string(indexData)

		children, err := os.ReadDir(sectionDir)
		if err != nil {
			t.Fatalf("read %s: %v", sectionDir, err)
		}

		var missing []string
		for _, child := range children {
			if child.IsDir() || filepath.Ext(child.Name()) != ".md" || child.Name() == "index.md" {
				continue
			}
			if !regexp.MustCompile(regexp.QuoteMeta(child.Name())).MatchString(indexText) {
				missing = append(missing, child.Name())
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			failures = append(failures, entry.Name()+":"+strings.Join(missing, ","))
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Fatalf("docs section index drift: %v", failures)
	}
}

func checkNoStaleConfirmExample(t *testing.T, path string, re *regexp.Regexp) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if re.Find(data) != nil {
		t.Fatalf("%s contains stale shared-confirm example missing colon separator", path)
	}
}

func TestP2PDocsExplainProvenanceGatewayAddressContract(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "docs", "cli", "p2p.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"`--addr` overrides the configured gateway address",
		"uses configured `server.host` and `server.port`",
		"Explicit `--addr` values are",
		"normalized before gateway requests",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("%s is missing P2P provenance gateway address snippet %q", target, snippet)
		}
	}
}

func docsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
