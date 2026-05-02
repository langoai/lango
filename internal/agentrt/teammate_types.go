package agentrt

import (
	"fmt"

	"github.com/langoai/lango/internal/toolchain"
)

type TeammateType struct {
	Name        string
	Description string
	MaxTools    map[string]bool
}

func (t TeammateType) AllowsTool(toolName string) bool {
	return t.MaxTools[toolName]
}

func BuiltinTeammateTypes() map[string]TeammateType {
	runtimeEssentials := append(toolchain.RuntimeEssentialToolNames(),
		"agent_spawn",
		"agent_wait",
		"agent_stop",
		"task_create",
		"task_get",
		"task_list",
		"task_update",
	)

	withRuntime := func(names ...string) map[string]bool {
		all := make([]string, 0, len(runtimeEssentials)+len(names))
		all = append(all, runtimeEssentials...)
		all = append(all, names...)
		return allowTools(all...)
	}

	return map[string]TeammateType{
		"operator": {
			Name:        "operator",
			Description: "Local execution and filesystem specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"exec",
				"exec_bg",
				"exec_status",
				"exec_stop",
				"fs_read",
				"fs_write",
				"fs_list",
				"fs_delete",
				"fs_edit",
				"fs_mkdir",
				"fs_stat",
			),
		},
		"navigator": {
			Name:        "navigator",
			Description: "Browser and web navigation specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"browser_navigate",
				"browser_search",
				"browser_observe",
				"browser_action",
				"browser_extract",
				"browser_screenshot",
				"web_search",
				"web_fetch",
			),
		},
		"vault": {
			Name:        "vault",
			Description: "Secrets, crypto, and payment specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"secrets_store",
				"secrets_get",
				"secrets_list",
				"secrets_delete",
				"crypto_encrypt",
				"crypto_decrypt",
				"crypto_sign",
				"crypto_hash",
				"crypto_keys",
				"payment_send",
				"payment_balance",
				"payment_history",
				"payment_limits",
				"payment_wallet_info",
				"payment_create_wallet",
				"payment_x402_fetch",
				"approve_upfront_payment",
				"execute_escrow_recommendation",
				"create_dispute_ready_receipt",
				"open_knowledge_exchange_transaction",
				"select_knowledge_exchange_path",
				"apply_settlement_progression",
				"adjudicate_escrow_dispute",
				"list_dead_lettered_post_adjudication_executions",
				"get_post_adjudication_execution_status",
				"retry_post_adjudication_execution",
				"execute_settlement",
				"execute_partial_settlement",
				"hold_escrow_for_dispute",
				"release_escrow_settlement",
				"refund_escrow_settlement",
			),
		},
		"librarian": {
			Name:        "librarian",
			Description: "Knowledge, learning, and skill specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"graph_query",
				"graph_traverse",
				"librarian_pending_inquiries",
				"librarian_dismiss_inquiry",
				"save_knowledge",
				"evaluate_exportability",
				"approve_artifact_release",
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
			),
		},
		"automator": {
			Name:        "automator",
			Description: "Background, cron, and workflow specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"bg_submit",
				"bg_status",
				"bg_list",
				"bg_result",
				"bg_cancel",
				"cron_add",
				"cron_list",
				"cron_pause",
				"cron_resume",
				"cron_remove",
				"cron_history",
				"workflow_run",
				"workflow_status",
				"workflow_list",
				"workflow_cancel",
				"workflow_save",
			),
		},
		"planner": {
			Name:        "planner",
			Description: "Planning and coordination specialist limited to runtime control essentials.",
			MaxTools:    allowTools(runtimeEssentials...),
		},
		"chronicler": {
			Name:        "chronicler",
			Description: "Memory specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"memory_list_observations",
				"memory_list_reflections",
				"memory_agent_save",
				"memory_agent_recall",
				"memory_agent_forget",
			),
		},
		"ontologist": {
			Name:        "ontologist",
			Description: "Ontology and schema specialist with runtime control essentials.",
			MaxTools: withRuntime(
				"ontology_list_actions",
				"ontology_action_link_entities",
				"ontology_action_set_entity_status",
				"ontology_promote_type",
				"ontology_promote_predicate",
				"ontology_schema_health",
				"ontology_type_usage",
				"ontology_list_types",
				"ontology_describe_type",
				"ontology_query_entities",
				"ontology_get_entity",
				"ontology_assert_fact",
				"ontology_retract_fact",
				"ontology_list_conflicts",
				"ontology_resolve_conflict",
				"ontology_merge_entities",
				"ontology_facts_at",
				"ontology_import_json",
				"ontology_import_csv",
				"ontology_from_mcp",
			),
		},
	}
}

func ValidateAllowedToolsForTeammate(teammateType string, allowedTools []string) error {
	if teammateType == "" || len(allowedTools) == 0 {
		return nil
	}

	builtin, ok := BuiltinTeammateTypes()[teammateType]
	if !ok {
		// Preserve Task 3 compatibility for non-built-in or legacy advisory agent names.
		return nil
	}

	for _, toolName := range allowedTools {
		if !builtin.AllowsTool(toolName) {
			return fmt.Errorf("tool %q outside role maximum scope for teammate type %q", toolName, teammateType)
		}
	}

	return nil
}

func allowTools(names ...string) map[string]bool {
	allowed := make(map[string]bool, len(names))
	for _, name := range names {
		allowed[name] = true
	}
	return allowed
}
