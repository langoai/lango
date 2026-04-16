package toolchain

import (
	"context"
	"fmt"
)

// DefaultSaveableTools is the app-level default set of tool names whose results
// are eligible for knowledge saving. Read-type tools only.
// P1: replace with tool catalog capability-based auto computation.
var DefaultSaveableTools = []string{
	"web_fetch",
	"web_search",
	"search_knowledge",
	"search_learnings",
	"get_knowledge_history",
	"graph_query",
	"graph_traverse",
	"rag_retrieve",
	"memory_list_observations",
	"memory_list_reflections",
	"memory_agent_recall",
	"browser_observe",
	"browser_extract",
	"ontology_query_entities",
	"ontology_get_entity",
	"ontology_describe_type",
	"ontology_list_types",
	"ontology_facts_at",
	"librarian_pending_inquiries",
	"learning_stats",
	"bg_status",
	"bg_list",
	"bg_result",
}

// KnowledgeSaver is the interface for saving tool results as knowledge.
// This avoids a direct import of the knowledge package.
type KnowledgeSaver interface {
	SaveToolResult(ctx context.Context, sessionKey, toolName string, params map[string]interface{}, result interface{}) error
}

// KnowledgeSaveHook auto-saves tool results as knowledge entries.
// Priority: 100 (runs last — after all other post-hooks).
type KnowledgeSaveHook struct {
	saver KnowledgeSaver

	// SaveableTools is the set of tool names whose results should be saved.
	// If empty, no results are saved (opt-in, not opt-out).
	SaveableTools map[string]bool
}

// Compile-time interface check.
var _ PostToolHook = (*KnowledgeSaveHook)(nil)

// NewKnowledgeSaveHook creates a new KnowledgeSaveHook.
func NewKnowledgeSaveHook(saver KnowledgeSaver, saveableTools []string) *KnowledgeSaveHook {
	m := make(map[string]bool, len(saveableTools))
	for _, t := range saveableTools {
		m[t] = true
	}
	return &KnowledgeSaveHook{saver: saver, SaveableTools: m}
}

// Name returns the hook name.
func (h *KnowledgeSaveHook) Name() string { return "knowledge_save" }

// Priority returns 100 (low priority — runs last).
func (h *KnowledgeSaveHook) Priority() int { return 100 }

// Post saves the tool result as knowledge if the tool is in the saveable set
// and the tool succeeded.
func (h *KnowledgeSaveHook) Post(ctx HookContext, result interface{}, toolErr error) error {
	if h.saver == nil || toolErr != nil {
		return nil
	}
	if !h.SaveableTools[ctx.ToolName] {
		return nil
	}

	if err := h.saver.SaveToolResult(ctx.Ctx, ctx.SessionKey, ctx.ToolName, ctx.Params, result); err != nil {
		return fmt.Errorf("knowledge save hook: %w", err)
	}
	return nil
}
