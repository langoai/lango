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
		filepath.Join(repoRoot, "README.md"),
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

func TestPublicDocsExplainBackgroundCLIServerBoundary(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "cli", "index.md"),
		filepath.Join(repoRoot, "docs", "automation", "background.md"),
	}
	requiredSnippets := []string{
		"Background task state is in-memory and owned by the running app/server process",
		"The current root CLI `lango bg` surface is not yet a remote gateway client",
		"in-app/cockpit task surfaces or agent `bg_*` tools",
		"until a remote management API exists",
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
				t.Fatalf("%s is missing background CLI server-boundary caveat snippet %q", target, snippet)
			}
		}
	}
}

func TestPublicDocsExplainBareRootNonInteractiveFallback(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "README.md"),
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

func TestArchitectureProjectStructureIncludesCurrentConfigCLISurface(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)

	requiredArchitectureSnippets := []string{
		"`cli/configcmd/`",
		"`lango config list`, `create`, `use`, `delete`, `import`, `export`, `get`, `set`, `keys`, `validate`",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe current config CLI inventory snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango config list/create/use/delete/import/export/get/set/keys/validate")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current config inventory order", readmeTarget)
	}
}

func TestArchitectureAndREADMEIncludeSharedCLISupportPackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`cli/cliboot/`",
		"Shared bootstrap loaders",
		"`BootResult` / `Config` callbacks",
		"`cli/clihttp/`",
		"Shared HTTP/JSON helpers for gateway-backed CLI commands",
		"`table|json` output validation",
		"`cli/workbenchstart/`",
		"starter, post-turn, and recovery prompt builders",
		"git branch/dirty state",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe shared CLI support snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"cliboot/        #   Shared CLI bootstrap / lazy config loading",
		"clihttp/        #   Shared HTTP/JSON helpers for gateway-backed CLI commands",
		"workbenchstart/ #   Context-aware starter/recovery prompts for bare lango",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe shared CLI support snippet %q", readmeTarget, snippet)
		}
	}
}

func TestREADMEIncludesCurrentMissionProjectionPackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"proposal/           # Transient proactive proposal registry, preparation, dismiss/accept flow",
		"loopview/           # Deterministic operator-loop and agenda projection from real runtime sources",
		"collabview/         # Deterministic mission-collaboration projection for local coworking state",
	}
	for _, snippet := range requiredSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(text) {
			t.Fatalf("%s does not describe mission projection package snippet %q", target, snippet)
		}
	}
}

func TestREADMEUsesCurrentGraphInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	if !regexp.MustCompile(regexp.QuoteMeta("lango graph status/query/stats/clear/add/export/import")).MatchString(text) {
		t.Fatalf("%s does not describe the current graph inventory", target)
	}
}

func TestArchitectureAndREADMEUseCurrentPaymentAndMetricsInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	if !regexp.MustCompile(regexp.QuoteMeta("`lango payment balance`, `history`, `limits`, `info`, `send`, `x402`")).MatchString(architectureText) {
		t.Fatalf("%s does not describe the current payment CLI surface", architectureTarget)
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango metrics/sessions/tools/agents/policy/history")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current metrics inventory", readmeTarget)
	}
	if !regexp.MustCompile(regexp.QuoteMeta("lango payment balance/history/limits/info/send/x402")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current payment inventory", readmeTarget)
	}
}

func TestArchitectureAndREADMEUseCurrentP2PInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`workspace create/list/status/join/leave`",
		"`git init/log/diff/push/fetch`",
		"`provenance push/fetch`",
		"`team list/status/disband`",
		"`zkp status/circuits`",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe current P2P inventory snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango p2p status/peers/connect/disconnect/firewall list/add/remove/discover/identity/reputation/pricing/session list/revoke/revoke-all/sandbox status/test/cleanup/workspace create/list/status/join/leave/git init/log/diff/push/fetch/provenance push/fetch/team list/status/disband/zkp status/circuits")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current P2P inventory", readmeTarget)
	}
}

func TestREADMEIncludesCurrentP2PPackageSubtree(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"p2p/                # P2P networking, collaborative workspaces, git/provenance exchange, trust policy, payments, and ZK proofs",
		"agentpool/      #   Agent pool with health checking and weighted selection",
		"discovery/      #   GossipSub agent card propagation, credential revocation",
		"firewall/       #   Default deny-all ACL with per-peer, per-tool rules",
		"gitbundle/      #   Incremental git bundle exchange for collaborative workspaces",
		"handshake/      #   ZK-enhanced authentication, session store, nonce cache",
		"identity/       #   DID identity derivation (did:lango:<pubkey>)",
		"ontologybridge/ #   Ontology/fact bridging across P2P peers",
		"paygate/        #   Payment gate, ledger, trust-based pricing",
		"protocol/       #   P2P protocol handler, remote agent, message types",
		"provenanceproto/#   Signed provenance bundle exchange protocol",
		"reputation/     #   Trust score tracking based on exchange outcomes",
		"settlement/     #   On-chain USDC settlement (EIP-3009)",
		"team/           #   P2P team coordination with conflict resolution",
		"trustpolicy/    #   Policy layer for peer trust and delegation constraints",
		"workspace/      #   Collaborative workspace runtime and membership state",
		"zkp/            #   ZK proofs (Plonk/Groth16), circuits (ownership, balance, capability, attestation)",
	}
	for _, snippet := range requiredSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(text) {
			t.Fatalf("%s does not describe current P2P package subtree snippet %q", target, snippet)
		}
	}
}

func TestArchitectureAndREADMEUseCurrentAlertsInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	if !regexp.MustCompile(regexp.QuoteMeta("`lango alerts list`, `summary`")).MatchString(architectureText) {
		t.Fatalf("%s does not describe the current alerts CLI surface", architectureTarget)
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango alerts list/summary")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current alerts inventory", readmeTarget)
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

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango memory list/status/clear/agents/agent <name>")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current memory inventory", readmeTarget)
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

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango contract read/call/abi load")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current contract inventory", readmeTarget)
	}
}

func TestArchitectureAndREADMEUseCurrentEconomyInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`lango economy budget status`",
		"`risk status`",
		"`pricing status`",
		"`negotiate status`",
		"`escrow status/list/show/sentinel status`",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe current economy inventory snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	if !regexp.MustCompile(regexp.QuoteMeta("lango economy budget status/risk status/pricing status/negotiate status/escrow status/list/show/sentinel status")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current economy inventory", readmeTarget)
	}
}

func TestREADMEUsesCurrentSecurityInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	text := string(data)

	requiredSnippets := []string{
		"lango security status/change-passphrase/deprecated migrate-passphrase/secrets/",
		"keyring store/clear/status",
		"recovery setup/restore",
		"kms status/test/keys/wrap/detach",
		"(+ legacy db-* tombstones)",
	}
	for _, snippet := range requiredSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(text) {
			t.Fatalf("%s does not describe current security inventory snippet %q", target, snippet)
		}
	}
}

func TestArchitectureAndREADMEUseCurrentRemainingCLIInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`lango chat` -- focused chat TUI",
		"`lango extension inspect/install/list/remove`",
		"`lango provenance status/checkpoint list/create/show/session tree/list/attribution show/report/bundle export/import`",
		"`lango run list/status/journal <run-id>`",
		"`lango sandbox status/test`",
		"`lango status`, `dead-letter-summary`, `dead-letters`, `dead-letter`, `dead-letter retry`",
		"`lango workflow run`, `list`, `status`, `cancel`, `history`, `validate <file>`",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe current CLI inventory snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"lango chat (focused chat TUI)",
		"lango extension inspect/install/list/remove",
		"lango provenance status/checkpoint list/create/show/session tree/list/attribution show/report/bundle export/import",
		"lango run list/status/journal <run-id>",
		"lango sandbox status/test",
		"lango status/dead-letter-summary/dead-letters/dead-letter/dead-letter retry",
		"lango workflow run/list/status/cancel/history/validate <file>",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe current README inventory snippet %q", readmeTarget, snippet)
		}
	}
}

func TestArchitectureAndREADMEUseCurrentA2AAndAgentInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`lango a2a card`, `lango a2a check`",
		"`lango agent status`, `list`, `tools`, `hooks`, `trace list/show/metrics`, `graph`",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe current inventory snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	a2aRow := "│   │   ├── a2a/            #   lango a2a card/check"
	if !regexp.MustCompile(regexp.QuoteMeta("lango a2a card/check")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current A2A inventory", readmeTarget)
	}
	if occurrences := regexp.MustCompile(regexp.QuoteMeta(a2aRow)).FindAllStringIndex(readmeText, -1); len(occurrences) != 1 {
		t.Fatalf("%s must contain exactly one A2A inventory row, found %d", readmeTarget, len(occurrences))
	}
	if !regexp.MustCompile(regexp.QuoteMeta("lango agent status/list/tools/hooks/trace list/show/metrics/graph")).MatchString(readmeText) {
		t.Fatalf("%s does not describe the current agent inventory", readmeTarget)
	}
	if regexp.MustCompile(regexp.QuoteMeta("lango chat (plain TUI chat)")).MatchString(readmeText) {
		t.Fatalf("%s still contains stale duplicate chat inventory wording", readmeTarget)
	}
}

func TestArchitectureAndREADMEIncludeRuntimeSupportPackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`exportability/`",
		"Source-class exportability policy evaluator",
		"`knowledgeruntime/`",
		"selects the execution branch (`prepay` or `escrow`)",
		"`receipts/`",
		"Canonical in-memory submission/transaction receipt store",
		"`storagebroker/`",
		"Persistent stdio JSON broker protocol for encrypted storage operations",
		"`streamx/`",
		"Generic iterator-based stream combinator package",
		"`tooloutput/`",
		"TTL-backed in-memory tool output store",
		"`toolparam/`",
		"Typed dynamic tool parameter extraction helpers",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe runtime support package snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"exportability/      # Source-class exportability evaluator producing receipt and lineage decisions",
		"knowledgeruntime/   # Knowledge-exchange runtime branch selector over canonical receipts and payment approval",
		"receipts/           # Canonical submission/transaction receipt store, events, settlement/runtime progression",
		"storagebroker/      # Persistent stdio JSON broker protocol for encrypted DB/config/session operations",
		"streamx/            # Generic iterator-based stream combinators with merge/race/fan-in helpers",
		"tooloutput/         # TTL-backed tool output store with ranged retrieval and regex grep helpers",
		"toolparam/          # Typed tool parameter extraction helpers for dynamic tool calls",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe runtime support package snippet %q", readmeTarget, snippet)
		}
	}
}

func TestArchitectureAndREADMEIncludePaymentSettlementSupportPackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`finance/`",
		"Shared monetary leaf utilities for USDC operations",
		"`paymentapproval/`",
		"Upfront-payment policy evaluator",
		"`paymentgate/`",
		"Direct-payment eligibility gate over canonical receipts",
		"`settlementprogression/`",
		"Canonical settlement progression mapper",
		"`settlementexecution/`",
		"Direct-payment settlement executor",
		"`partialsettlementexecution/`",
		"Partial direct-payment settlement executor",
		"`escrowexecution/`",
		"Escrow create/fund runtime bridge",
		"`disputehold/`",
		"Dispute-hold executor for funded escrow transactions",
		"`escrowadjudication/`",
		"Canonical escrow adjudication applier",
		"`escrowrelease/`",
		"Escrow release executor for funded, release-adjudicated transactions",
		"`escrowrefund/`",
		"Escrow refund executor for funded, refund-adjudicated transactions",
		"`postadjudicationreplay/`",
		"Manual post-adjudication replay dispatcher",
		"`postadjudicationstatus/`",
		"Dead-letter and retry-status projection over adjudicated transactions",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe payment-settlement support snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"finance/            # Shared USDC parsing, formatting, and quote helpers as a leaf monetary utility package",
		"paymentapproval/    # Upfront-payment policy evaluator producing approve/reject/escalate and prepay/escrow hints",
		"paymentgate/        # Direct-payment eligibility gate over canonical receipts and current submission bindings",
		"settlementprogression/ # Release-outcome to settlement/dispute progression mapper over canonical receipts",
		"settlementexecution/ # Direct-payment settlement executor for approved canonical transactions",
		"partialsettlementexecution/ # Partial direct-payment executor with executed/remaining amount tracking",
		"escrowexecution/    # Escrow create/fund runtime bridge for escrow-recommended approved transactions",
		"disputehold/        # Dispute-hold executor for funded escrow transactions in dispute-ready state",
		"escrowadjudication/ # Canonical escrow adjudication applier for release/refund outcomes after hold evidence",
		"escrowrelease/      # Escrow release executor for release-adjudicated funded transactions",
		"escrowrefund/       # Escrow refund executor for refund-adjudicated funded transactions",
		"postadjudicationreplay/ # Manual post-adjudication retry dispatcher gated by dead-letter evidence and actor policy",
		"postadjudicationstatus/ # Dead-letter backlog and retry-status projection over canonical adjudication state",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe payment-settlement support snippet %q", readmeTarget, snippet)
		}
	}
}

func TestArchitectureAndREADMEIncludeExecutionRetrievalInfrastructurePackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`agentrt/`",
		"Agent runtime control-plane package",
		"`gatekeeper/`",
		"Response sanitization package",
		"`retrieval/`",
		"Retrieval orchestration package",
		"`search/`",
		"Domain-agnostic FTS5 search substrate",
		"`turnrunner/`",
		"Shared turn execution runner",
		"`turntrace/`",
		"Durable turn trace package",
		"`lineio/`",
		"Shared single-line reader helper",
		"`storeutil/`",
		"Small store-facing utility helpers",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe execution/retrieval infrastructure snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"agentrt/            # Agent runtime control plane (coordinating executor, delegation guard, budget, recovery)",
		"gatekeeper/         # Response sanitization (thought tags, internal markers, raw JSON, custom patterns)",
		"retrieval/          # Retrieval coordinator with fact/temporal agents, dedup, reranking, token-budget truncation",
		"search/             # FTS5 search substrate (domain-agnostic full-text CRUD)",
		"turnrunner/         # Shared turn execution runner with timeout, retry, tracing, delegation/tool/thinking callbacks",
		"turntrace/          # Durable turn trace store, append-only events, failure queries, and per-agent metrics",
		"lineio/             # Shared single-line reader preserving partial line + EOF behavior",
		"storeutil/          # Generic slice/map copy and JSON marshal/unmarshal helpers for stores",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe execution/retrieval infrastructure snippet %q", readmeTarget, snippet)
		}
	}
}

func TestArchitectureAndREADMEIncludeOperationalSupportPackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`alerting/`",
		"Operational alerting package",
		"`approvalflow/`",
		"Canonical artifact release approval-flow package",
		"`archtest/`",
		"Architecture enforcement test package",
		"`dbopen/`",
		"Managed database-opening helpers",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe operational support snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"approvalflow/       # Canonical artifact release decision mapper over exportability, risk, and settlement hints",
		"alerting/           # Threshold-based operational alert dispatcher and webhook delivery fan-out",
		"archtest/           # Architecture boundary and bootstrap/storage wiring enforcement tests",
		"dbopen/             # Managed/read-only Ent+SQLite open helpers with serialized schema migration and connection setup",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe operational support snippet %q", readmeTarget, snippet)
		}
	}
}

func TestArchitectureAndREADMEIncludeOntologyStoragePackages(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)
	requiredArchitectureSnippets := []string{
		"`ontology/`",
		"Ontology governance and tooling package",
		"`sqlitedriver/`",
		"Shared SQLite driver helper package",
		"`storage/`",
		"Storage facade and broker-adapter package",
	}
	for _, snippet := range requiredArchitectureSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(architectureText) {
			t.Fatalf("%s does not describe ontology/storage snippet %q", architectureTarget, snippet)
		}
	}

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)
	requiredReadmeSnippets := []string{
		"ontology/           # Ontology registry, ACL/action/truth governance, property/exchange services, Ent-backed stores",
		"sqlitedriver/       # Shared SQLite open/config/header-check helpers for managed and read-only DB access",
		"storage/            # Storage facade composing config/security/session/workflow/ontology/payment persistence and broker-backed adapters",
	}
	for _, snippet := range requiredReadmeSnippets {
		if !regexp.MustCompile(regexp.QuoteMeta(snippet)).MatchString(readmeText) {
			t.Fatalf("%s does not describe ontology/storage snippet %q", readmeTarget, snippet)
		}
	}
}

func TestTopLevelInternalPackagesAppearInREADMEAndArchitectureInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)

	internalDir := filepath.Join(repoRoot, "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		t.Fatalf("read %s: %v", internalDir, err)
	}

	var missingReadme []string
	var missingArchitecture []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		snippet := regexp.QuoteMeta(name + "/")
		if !regexp.MustCompile(snippet).MatchString(readmeText) {
			missingReadme = append(missingReadme, name)
		}
		if !regexp.MustCompile(snippet).MatchString(architectureText) {
			missingArchitecture = append(missingArchitecture, name)
		}
	}

	sort.Strings(missingReadme)
	sort.Strings(missingArchitecture)

	if len(missingReadme) > 0 || len(missingArchitecture) > 0 {
		t.Fatalf(
			"top-level internal package inventory drift: missing in README=%v missing in architecture=%v",
			missingReadme,
			missingArchitecture,
		)
	}
}

func TestCLIInternalPackagesAppearInREADMEAndArchitectureInventory(t *testing.T) {
	t.Parallel()

	repoRoot := docsQualityRepoRoot(t)

	readmeTarget := filepath.Join(repoRoot, "README.md")
	readmeData, err := os.ReadFile(readmeTarget)
	if err != nil {
		t.Fatalf("read %s: %v", readmeTarget, err)
	}
	readmeText := string(readmeData)

	architectureTarget := filepath.Join(repoRoot, "docs", "architecture", "project-structure.md")
	architectureData, err := os.ReadFile(architectureTarget)
	if err != nil {
		t.Fatalf("read %s: %v", architectureTarget, err)
	}
	architectureText := string(architectureData)

	cliDir := filepath.Join(repoRoot, "internal", "cli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		t.Fatalf("read %s: %v", cliDir, err)
	}

	var missingReadme []string
	var missingArchitecture []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !regexp.MustCompile(regexp.QuoteMeta("├── " + name + "/")).MatchString(readmeText) {
			missingReadme = append(missingReadme, name)
		}
		if !regexp.MustCompile(regexp.QuoteMeta("cli/" + name + "/")).MatchString(architectureText) {
			missingArchitecture = append(missingArchitecture, name)
		}
	}

	sort.Strings(missingReadme)
	sort.Strings(missingArchitecture)

	if len(missingReadme) > 0 || len(missingArchitecture) > 0 {
		t.Fatalf(
			"internal/cli package inventory drift: missing in README=%v missing in architecture=%v",
			missingReadme,
			missingArchitecture,
		)
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

func docsQualityRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
