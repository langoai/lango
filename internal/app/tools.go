package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/automation"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/tools/browser"
	execpkg "github.com/langoai/lango/internal/tools/exec"
	"github.com/langoai/lango/internal/tools/filesystem"
	"github.com/langoai/lango/internal/tools/webfetch"
	"github.com/langoai/lango/internal/tools/websearch"
)

// buildTools creates the set of tools available to the agent.
// When browserSM is non-nil, browser tools are included.
// automationAvailable indicates which automation features are enabled (cron, background, workflow).
func buildTools(sv *supervisor.Supervisor, fsCfg filesystem.Config, browserSM *browser.SessionManager, automationAvailable map[string]bool, guard *execpkg.CommandGuard) []*agent.Tool {
	var tools []*agent.Tool

	// Exec tools (delegated to Supervisor for security isolation).
	// Guard functions stay in app — they depend on app-level knowledge.
	langoGuard := func(cmd string) string { return blockLangoExec(cmd, automationAvailable) }
	pathGuard := func(cmd string) string { return blockProtectedPaths(cmd, guard) }
	tools = append(tools, execpkg.BuildTools(sv, langoGuard, pathGuard)...)

	// Filesystem tools
	fsTool := filesystem.New(fsCfg)
	tools = append(tools, filesystem.BuildTools(fsTool)...)

	// Browser tools (opt-in), wrapped with panic recovery
	if browserSM != nil {
		for _, bt := range browser.BuildTools(browserSM) {
			tools = append(tools, wrapBrowserHandler(bt, browserSM))
		}
	}

	// Web search and fetch tools (HTTP-only, no browser required)
	tools = append(tools, websearch.BuildTools()...)
	tools = append(tools, webfetch.BuildTools()...)

	return tools
}

// classifyLangoExec checks if the command attempts to invoke the lango CLI
// or redirects skill-related commands. Returns a guidance message and a
// structured ReasonCode for the PolicyEvaluator.
func classifyLangoExec(cmd string, automationAvailable map[string]bool) (string, execpkg.ReasonCode) {
	lower := strings.ToLower(strings.TrimSpace(cmd))

	// --- Phase 1: Subcommands with in-process tool equivalents ---
	type guard struct {
		prefix  string
		feature string // key in automationAvailable; empty = always available
		tools   string
	}
	guards := []guard{
		{"lango cron", "cron", "cron_add, cron_list, cron_pause, cron_resume, cron_remove, cron_history"},
		{"lango bg", "background", "bg_submit, bg_status, bg_list, bg_result, bg_cancel"},
		{"lango background", "background", "bg_submit, bg_status, bg_list, bg_result, bg_cancel"},
		{"lango workflow", "workflow", "workflow_run, workflow_status, workflow_list, workflow_cancel, workflow_save"},
		{"lango graph", "", "graph_traverse, graph_query, rag_retrieve"},
		{"lango memory", "", "memory_list_observations, memory_list_reflections"},
		{"lango p2p", "", "p2p_status, p2p_connect, p2p_disconnect, p2p_peers, p2p_query, p2p_discover, p2p_firewall_rules, p2p_firewall_add, p2p_firewall_remove, p2p_reputation, p2p_pay, p2p_price_query"},
		{"lango security", "", "crypto_encrypt, crypto_decrypt, crypto_sign, crypto_hash, crypto_keys, secrets_store, secrets_get, secrets_list, secrets_delete"},
		{"lango payment", "", "payment_send, payment_create_wallet, payment_x402_fetch"},
		{"lango mcp", "", "mcp_status, mcp_tools"},
		{"lango contract", "", "contract_read, contract_call, contract_abi_load"},
		{"lango account", "", "smart_account_deploy, smart_account_info, session_key_create, session_key_list, session_key_revoke, session_execute, policy_check, module_install, module_uninstall, spending_status, paymaster_status, paymaster_approve"},
	}

	for _, g := range guards {
		if strings.HasPrefix(lower, g.prefix) {
			if g.feature == "" || automationAvailable[g.feature] {
				return fmt.Sprintf(
					"Do not use exec to run '%s' — use the built-in tools instead (%s). "+
						"Spawning a new lango process requires passphrase authentication and will fail in non-interactive mode.",
					g.prefix, g.tools), execpkg.ReasonLangoCLI
			}
			return fmt.Sprintf(
				"Cannot run '%s' via exec — spawning a new lango process requires passphrase authentication. "+
					"Enable the %s feature in Settings to use the built-in tools (%s).",
				g.prefix, g.feature, g.tools), execpkg.ReasonLangoCLI
		}
	}

	// --- Phase 2: Catch-all for any remaining lango subcommand ---
	if strings.HasPrefix(lower, "lango ") || lower == "lango" {
		return "Do not use exec to run the lango CLI — every lango command requires passphrase authentication " +
			"via bootstrap and will fail when spawned as a subprocess. " +
			"Use the built-in tools (try builtin_list to discover available tools), " +
			"or ask the user to run this command directly in their terminal.", execpkg.ReasonLangoCLI
	}

	// --- Phase 3: Skill import redirects ---

	// Redirect skill-related git clone to import_skill tool.
	if strings.HasPrefix(lower, "git clone") && strings.Contains(lower, "skill") {
		return "Use the built-in import_skill tool instead of manual git clone — " +
			"it automatically uses git clone internally when available and stores skills in the correct location (~/.lango/skills/). " +
			"Example: import_skill(url: \"<github-repo-url>\")", execpkg.ReasonSkillImport
	}

	// Redirect skill-related curl/wget to import_skill tool.
	if (strings.HasPrefix(lower, "curl ") || strings.HasPrefix(lower, "wget ")) &&
		strings.Contains(lower, "skill") {
		return "Use the built-in import_skill tool instead of manual curl/wget — " +
			"it handles downloads internally and stores skills correctly. " +
			"Example: import_skill(url: \"<url>\")", execpkg.ReasonSkillImport
	}

	return "", execpkg.ReasonNone
}

// blockLangoExec is the GuardFunc-compatible wrapper for handler-level
// defense-in-depth. Delegates to classifyLangoExec and returns only the
// message string.
func blockLangoExec(cmd string, automationAvailable map[string]bool) string {
	msg, _ := classifyLangoExec(cmd, automationAvailable)
	return msg
}

// blockProtectedPaths checks if the command attempts to access protected data
// paths or execute dangerous process management commands.
// Returns a guidance message if blocked, or empty string if allowed.
func blockProtectedPaths(cmd string, guard *execpkg.CommandGuard) string {
	if guard == nil {
		return ""
	}
	blocked, reason := guard.CheckCommand(cmd)
	if blocked {
		return reason
	}
	return ""
}

// wrapBrowserHandler wraps a browser tool handler with panic recovery and auto-reconnect.
// Delegates to toolchain.WithBrowserRecovery.
func wrapBrowserHandler(t *agent.Tool, sm *browser.SessionManager) *agent.Tool {
	return toolchain.Chain(t, toolchain.WithBrowserRecovery(sm))
}

// detectChannelFromContext delegates to automation.DetectChannelFromContext.
func detectChannelFromContext(ctx context.Context) string {
	return automation.DetectChannelFromContext(ctx)
}

// needsApproval delegates to toolchain.NeedsApproval.
func needsApproval(t *agent.Tool, ic config.InterceptorConfig) bool {
	return toolchain.NeedsApproval(t, ic)
}

// buildApprovalSummary delegates to toolchain.BuildApprovalSummary.
func buildApprovalSummary(toolName string, params map[string]interface{}) string {
	return toolchain.BuildApprovalSummary(toolName, params)
}

// truncate delegates to toolchain.Truncate.
func truncate(s string, maxLen int) string {
	return toolchain.Truncate(s, maxLen)
}
