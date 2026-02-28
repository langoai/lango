package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/tools/browser"
	"github.com/langoai/lango/internal/tools/filesystem"
	"github.com/langoai/lango/internal/types"
	"github.com/langoai/lango/internal/wallet"
)

// buildTools creates the set of tools available to the agent.
// When browserSM is non-nil, browser tools are included.
// automationAvailable indicates which automation features are enabled (cron, background, workflow).
func buildTools(sv *supervisor.Supervisor, fsCfg filesystem.Config, browserSM *browser.SessionManager, automationAvailable map[string]bool) []*agent.Tool {
	var tools []*agent.Tool

	// Exec tools (delegated to Supervisor for security isolation)
	tools = append(tools, buildExecTools(sv, automationAvailable)...)

	// Filesystem tools
	fsTool := filesystem.New(fsCfg)
	tools = append(tools, buildFilesystemTools(fsTool)...)

	// Browser tools (opt-in), wrapped with panic recovery
	if browserSM != nil {
		for _, bt := range buildBrowserTools(browserSM) {
			tools = append(tools, wrapBrowserHandler(bt, browserSM))
		}
	}

	return tools
}

// blockLangoExec checks if the command attempts to invoke the lango CLI.
// ALL lango CLI commands require passphrase authentication via bootstrap and
// will fail when spawned as a subprocess (non-interactive stdin). Returns a
// guidance message if blocked, or empty string if allowed.
func blockLangoExec(cmd string, automationAvailable map[string]bool) string {
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
	}

	for _, g := range guards {
		if strings.HasPrefix(lower, g.prefix) {
			if g.feature == "" || automationAvailable[g.feature] {
				return fmt.Sprintf(
					"Do not use exec to run '%s' — use the built-in tools instead (%s). "+
						"Spawning a new lango process requires passphrase authentication and will fail in non-interactive mode.",
					g.prefix, g.tools)
			}
			return fmt.Sprintf(
				"Cannot run '%s' via exec — spawning a new lango process requires passphrase authentication. "+
					"Enable the %s feature in Settings to use the built-in tools (%s).",
				g.prefix, g.feature, g.tools)
		}
	}

	// --- Phase 2: Catch-all for any remaining lango subcommand ---
	if strings.HasPrefix(lower, "lango ") || lower == "lango" {
		return "Do not use exec to run the lango CLI — every lango command requires passphrase authentication " +
			"via bootstrap and will fail when spawned as a subprocess. " +
			"Use the built-in tools for the operation you need, or ask the user to run this command directly in their terminal."
	}

	// Redirect skill-related git clone to import_skill tool.
	if strings.HasPrefix(lower, "git clone") && strings.Contains(lower, "skill") {
		return "Use the built-in import_skill tool instead of manual git clone — " +
			"it automatically uses git clone internally when available and stores skills in the correct location (~/.lango/skills/). " +
			"Example: import_skill(url: \"<github-repo-url>\")"
	}

	// Redirect skill-related curl/wget to import_skill tool.
	if (strings.HasPrefix(lower, "curl ") || strings.HasPrefix(lower, "wget ")) &&
		strings.Contains(lower, "skill") {
		return "Use the built-in import_skill tool instead of manual curl/wget — " +
			"it handles downloads internally and stores skills correctly. " +
			"Example: import_skill(url: \"<url>\")"
	}

	return ""
}

// wrapBrowserHandler wraps a browser tool handler with panic recovery and auto-reconnect.
// Delegates to toolchain.WithBrowserRecovery.
func wrapBrowserHandler(t *agent.Tool, sm *browser.SessionManager) *agent.Tool {
	return toolchain.Chain(t, toolchain.WithBrowserRecovery(sm))
}

// wrapWithLearning wraps a tool's handler to call the learning observer after each execution.
// Delegates to toolchain.WithLearning.
func wrapWithLearning(t *agent.Tool, observer learning.ToolResultObserver) *agent.Tool {
	return toolchain.Chain(t, toolchain.WithLearning(observer))
}

// detectChannelFromContext extracts the delivery target from the session key in context.
// Returns "channel:targetID" (e.g. "telegram:123456789") or "" if no known channel prefix is found.
func detectChannelFromContext(ctx context.Context) string {
	sessionKey := session.SessionKeyFromContext(ctx)
	if sessionKey == "" {
		return ""
	}
	// Session key format: "channel:targetID:userID"
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return ""
	}
	ch := types.ChannelType(parts[0])
	if ch.Valid() {
		return parts[0] + ":" + parts[1]
	}
	return ""
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

// wrapWithApproval wraps a tool to require approval based on the configured policy.
// Delegates to toolchain.WithApproval.
func wrapWithApproval(t *agent.Tool, ic config.InterceptorConfig, ap approval.Provider, gs *approval.GrantStore, limiter wallet.SpendingLimiter) *agent.Tool {
	return toolchain.Chain(t, toolchain.WithApproval(ic, ap, gs, limiter))
}
