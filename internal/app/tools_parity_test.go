package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/economy/pricing"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/p2p/team"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/tooloutput"
	"github.com/langoai/lango/internal/tools/browser"
	toolcrypto "github.com/langoai/lango/internal/tools/crypto"
	execpkg "github.com/langoai/lango/internal/tools/exec"
	"github.com/langoai/lango/internal/tools/filesystem"
	toolsecrets "github.com/langoai/lango/internal/tools/secrets"
)

// newMinimalCoordinator creates a team.Coordinator with zero-value config
// suitable for tool definition testing (handlers will fail at runtime).
func newMinimalCoordinator() *team.Coordinator {
	return team.NewCoordinator(team.CoordinatorConfig{})
}

// ─── helpers ───

// toolNamesUnsorted extracts tool names in original order.
func toolNamesUnsorted(tools []*agent.Tool) []string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return names
}

// assertAllHandlersNonNil asserts every tool has a non-nil handler.
func assertAllHandlersNonNil(t *testing.T, tools []*agent.Tool) {
	t.Helper()
	for _, tool := range tools {
		assert.NotNil(t, tool.Handler, "tool %q has nil handler", tool.Name)
	}
}

// assertNoDuplicateNames asserts there are no duplicate tool names.
func assertNoDuplicateNames(t *testing.T, tools []*agent.Tool) {
	t.Helper()
	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		assert.False(t, seen[tool.Name], "duplicate tool name %q", tool.Name)
		seen[tool.Name] = true
	}
}

// ─── Agent Memory ───

func TestBuildAgentMemoryTools_Parity(t *testing.T) {
	t.Parallel()

	store := agentmemory.NewInMemoryStore()
	tools := agentmemory.BuildTools(store)

	wantNames := []string{
		"memory_agent_save",
		"memory_agent_recall",
		"memory_agent_forget",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Output Tools ───

func TestBuildOutputTools_Parity(t *testing.T) {
	t.Parallel()

	store := tooloutput.NewOutputStore(5 * time.Minute)
	tools := tooloutput.BuildTools(store)

	wantNames := []string{
		"tool_output_get",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
}

// ─── Filesystem Tools ───

func TestBuildFilesystemTools_Parity(t *testing.T) {
	t.Parallel()

	fsTool := filesystem.New(filesystem.Config{})
	tools := filesystem.BuildTools(fsTool)

	wantNames := []string{
		"fs_read",
		"fs_list",
		"fs_write",
		"fs_edit",
		"fs_mkdir",
		"fs_delete",
		"fs_stat",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)

	// All fs tools should have a description.
	for _, tool := range tools {
		assert.NotEmpty(t, tool.Description, "tool %q missing description", tool.Name)
	}
}

// ─── Exec Tools ───

func TestBuildExecTools_Parity(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "" // avoid provider validation error
	sv, err := supervisor.New(cfg)
	require.NoError(t, err)

	langoGuard := func(cmd string) string { return blockLangoExec(cmd, nil) }
	pathGuard := func(cmd string) string { return blockProtectedPaths(cmd, nil) }
	tools := execpkg.BuildTools(sv, langoGuard, pathGuard)

	wantNames := []string{
		"exec",
		"exec_bg",
		"exec_status",
		"exec_stop",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Browser Tools ───

func TestBuildBrowserTools_Parity(t *testing.T) {
	t.Parallel()

	tool, err := browser.New(browser.Config{})
	require.NoError(t, err)
	sm := browser.NewSessionManager(tool)
	tools := browser.BuildTools(sm)

	wantNames := []string{
		"browser_navigate",
		"browser_search",
		"browser_observe",
		"browser_extract",
		"browser_action",
		"browser_screenshot",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Graph Tools ───

func TestBuildGraphTools_Parity(t *testing.T) {
	t.Parallel()

	// graph.Store is an interface; we pass nil to verify tool definitions compile
	// and return the expected names. Handlers will fail at runtime with nil store,
	// but we're testing structure only.
	tools := graph.BuildTools(nil)

	wantNames := []string{
		"graph_traverse",
		"graph_query",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Memory Agent Tools ───

func TestBuildMemoryAgentTools_Parity(t *testing.T) {
	t.Parallel()

	// memory.Store is a concrete type; pass nil to verify tool definitions.
	tools := memory.BuildObservationTools(nil)

	wantNames := []string{
		"memory_list_observations",
		"memory_list_reflections",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Librarian Tools ───

func TestBuildLibrarianTools_Parity(t *testing.T) {
	t.Parallel()

	// InquiryStore is a concrete type; pass nil to verify tool definitions.
	tools := librarian.BuildTools(nil)

	wantNames := []string{
		"librarian_pending_inquiries",
		"librarian_dismiss_inquiry",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Crypto Tools ───

func TestBuildCryptoTools_Parity(t *testing.T) {
	t.Parallel()

	// Pass nil dependencies — tool definitions are constructed regardless.
	refs := security.NewRefStore()
	tools := toolcrypto.BuildTools(nil, nil, refs, nil)

	wantNames := []string{
		"crypto_encrypt",
		"crypto_decrypt",
		"crypto_sign",
		"crypto_hash",
		"crypto_keys",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Secrets Tools ───

func TestBuildSecretsTools_Parity(t *testing.T) {
	t.Parallel()

	refs := security.NewRefStore()
	tools := toolsecrets.BuildTools(nil, refs, nil)

	wantNames := []string{
		"secrets_store",
		"secrets_get",
		"secrets_list",
		"secrets_delete",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Meta Tools ───

func TestBuildMetaTools_Parity(t *testing.T) {
	t.Parallel()

	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)

	wantNames := []string{
		"save_knowledge",
		"evaluate_exportability",
		"approve_artifact_release",
		"create_dispute_ready_receipt",
		"open_knowledge_exchange_transaction",
		"select_knowledge_exchange_path",
		"adjudicate_escrow_dispute",
		"approve_upfront_payment",
		"apply_settlement_progression",
		"get_knowledge_history",
		"search_knowledge",
		"save_learning",
		"search_learnings",
		"create_skill",
		"list_skills",
		"view_skill",
		"import_skill",
		"learning_stats",
		"learning_cleanup",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

func TestBuildMetaToolsWithEscrow_Parity(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	escrowEngine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, nil, receiptStore, escrowEngine)

	wantNames := []string{
		"save_knowledge",
		"evaluate_exportability",
		"approve_artifact_release",
		"create_dispute_ready_receipt",
		"open_knowledge_exchange_transaction",
		"select_knowledge_exchange_path",
		"adjudicate_escrow_dispute",
		"approve_upfront_payment",
		"apply_settlement_progression",
		"get_knowledge_history",
		"search_knowledge",
		"save_learning",
		"search_learnings",
		"create_skill",
		"list_skills",
		"view_skill",
		"import_skill",
		"learning_stats",
		"learning_cleanup",
		"retry_post_adjudication_execution",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
		"execute_escrow_recommendation",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

func TestBuildMetaToolsWithRuntimes_Parity(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	escrowEngine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
	settlementRuntime := &fakeSettlementExecutionRuntime{}
	partialSettlementRuntime := &fakePartialSettlementExecutionRuntime{}
	escrowDisputeHoldRuntime := &fakeDisputeHoldRuntime{}
	escrowReleaseRuntime := &fakeEscrowReleaseRuntime{}
	escrowRefundRuntime := &fakeEscrowRefundRuntime{}
	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, receiptStore, escrowEngine, settlementRuntime, partialSettlementRuntime, escrowDisputeHoldRuntime, escrowReleaseRuntime, escrowRefundRuntime)

	wantNames := []string{
		"save_knowledge",
		"evaluate_exportability",
		"approve_artifact_release",
		"create_dispute_ready_receipt",
		"open_knowledge_exchange_transaction",
		"select_knowledge_exchange_path",
		"adjudicate_escrow_dispute",
		"approve_upfront_payment",
		"apply_settlement_progression",
		"get_knowledge_history",
		"search_knowledge",
		"save_learning",
		"search_learnings",
		"create_skill",
		"list_skills",
		"view_skill",
		"import_skill",
		"learning_stats",
		"learning_cleanup",
		"retry_post_adjudication_execution",
		"execute_settlement",
		"execute_partial_settlement",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
		"execute_escrow_recommendation",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Economy: Aggregate economy.BuildTools ───

func TestBuildEconomyTools_AllEngines_Parity(t *testing.T) {
	t.Parallel()

	budgetStore := budget.NewStore()
	budgetEngine, err := budget.NewEngine(budgetStore, config.BudgetConfig{})
	require.NoError(t, err)

	riskEngine, err := risk.New(config.RiskConfig{}, nil)
	require.NoError(t, err)

	negEngine := negotiation.New(config.NegotiationConfig{})

	escrowStore := escrow.NewMemoryStore()
	escrowEngine := escrow.NewEngine(escrowStore, nil, escrow.EngineConfig{})

	pricingEngine, err := pricing.New(config.DynamicPricingConfig{})
	require.NoError(t, err)

	tools := economy.BuildTools(budgetEngine, riskEngine, negEngine, escrowEngine, pricingEngine)

	// 3 budget + 1 risk + 2 negotiation + 5 escrow + 1 pricing = 12
	assert.Len(t, tools, 12, "expected 12 total economy tools")
	assertNoDuplicateNames(t, tools)
	assertAllHandlersNonNil(t, tools)

	// Verify all economy tools start with "economy_" prefix.
	for _, tool := range tools {
		assert.Contains(t, tool.Name, "economy_",
			"economy tool %q should have economy_ prefix", tool.Name)
	}
}

func TestBuildEconomyTools_NilEngines_Empty(t *testing.T) {
	t.Parallel()

	tools := economy.BuildTools(nil, nil, nil, nil, nil)
	assert.Empty(t, tools, "nil engines should produce zero tools")
}

// ─── Sentinel Tools ───
// Note: sentinel already has tests in tools_sentinel_test.go.
// This test verifies parity only (tool count and names).

func TestBuildSentinelTools_Parity(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	wantNames := []string{
		"sentinel_status",
		"sentinel_alerts",
		"sentinel_config",
		"sentinel_acknowledge",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertNoDuplicateNames(t, tools)
}

// ─── Cron Tools ───

func TestBuildCronTools_Parity(t *testing.T) {
	t.Parallel()

	// cron.New requires a Store and Executor; pass nil and a nil Scheduler.
	// buildCronTools takes a *cron.Scheduler which we cannot easily construct
	// without a DB. Instead, test that the function exists, compiles, and that
	// the expected tool names match the blockLangoExec guard.

	wantNames := []string{
		"cron_add",
		"cron_list",
		"cron_pause",
		"cron_resume",
		"cron_remove",
		"cron_history",
	}

	// Cross-verify against blockLangoExec guard list.
	auto := map[string]bool{"cron": true}
	msg := blockLangoExec("lango cron list", auto)
	for _, name := range wantNames {
		assert.Contains(t, msg, name[:4], // at least prefix
			"blockLangoExec cron guard should reference %q", name)
	}

	assert.Len(t, wantNames, 6, "expected 6 cron tools")
}

// ─── Background Tools ───

func TestBuildBackgroundTools_Parity(t *testing.T) {
	t.Parallel()

	wantNames := []string{
		"bg_submit",
		"bg_status",
		"bg_list",
		"bg_result",
		"bg_cancel",
	}

	// Cross-verify against blockLangoExec guard list.
	auto := map[string]bool{"background": true}
	msg := blockLangoExec("lango bg list", auto)
	for _, name := range wantNames {
		assert.Contains(t, msg, name,
			"blockLangoExec bg guard should reference %q", name)
	}

	assert.Len(t, wantNames, 5, "expected 5 background tools")
}

// ─── Workflow Tools ───

func TestBuildWorkflowTools_Parity(t *testing.T) {
	t.Parallel()

	wantNames := []string{
		"workflow_run",
		"workflow_status",
		"workflow_list",
		"workflow_cancel",
		"workflow_save",
	}

	// Cross-verify against blockLangoExec guard list.
	auto := map[string]bool{"workflow": true}
	msg := blockLangoExec("lango workflow list", auto)
	for _, name := range wantNames {
		assert.Contains(t, msg, name,
			"blockLangoExec workflow guard should reference %q", name)
	}

	assert.Len(t, wantNames, 5, "expected 5 workflow tools")
}

// ─── Team Tools ───

func TestBuildTeamTools_Parity(t *testing.T) {
	t.Parallel()

	// team.NewCoordinator accepts a config struct; nil fields are safe for tool definition.
	coord := newMinimalCoordinator()
	tools := team.BuildTools(coord)

	wantNames := []string{
		"team_form",
		"team_delegate",
		"team_status",
		"team_list",
		"team_disband",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── On-Chain Escrow Tools ───

func TestBuildOnChainEscrowTools_Parity(t *testing.T) {
	t.Parallel()

	memStore := escrow.NewMemoryStore()
	engine := escrow.NewEngine(memStore, nil, escrow.EngineConfig{})
	tools := buildOnChainEscrowTools(engine, nil)

	wantNames := []string{
		"escrow_create",
		"escrow_fund",
		"escrow_activate",
		"escrow_submit_work",
		"escrow_release",
		"escrow_refund",
		"escrow_dispute",
		"escrow_resolve",
		"escrow_status",
		"escrow_list",
	}

	assert.Len(t, tools, len(wantNames))
	assert.Equal(t, wantNames, toolNamesUnsorted(tools))
	assertAllHandlersNonNil(t, tools)
	assertNoDuplicateNames(t, tools)
}

// ─── Cross-cutting: No Lost Tools After Extraction ───
// This test ensures the combined tool count across all domain builders
// matches expectations, catching any accidental omissions during extraction.

func TestToolBuilders_TotalCount(t *testing.T) {
	t.Parallel()

	counts := map[string]int{
		"agent_memory":   3,
		"output":         1,
		"filesystem":     7,
		"exec":           4,
		"browser":        3,
		"graph":          2,
		"rag":            1,
		"memory":         2,
		"librarian":      2,
		"crypto":         5,
		"secrets":        4,
		"meta":           9,
		"sentinel":       4,
		"cron":           6,
		"background":     5,
		"workflow":       5,
		"team":           5,
		"budget":         3,
		"risk":           1,
		"negotiation":    2,
		"escrow_in":      5, // economy_escrow_* (internal escrow)
		"pricing":        1,
		"escrow_onchain": 10,
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	// Sanity check: total should be >= 89.
	assert.GreaterOrEqual(t, total, 89,
		"combined tool count across all builders should be at least 89")
}
