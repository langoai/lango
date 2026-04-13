package runledger

import (
	"context"
	"fmt"
	"sync"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
)

var (
	orchestratorOnlyRunTools = map[string]struct{}{
		"run_create":       {},
		"run_apply_policy": {},
		"run_approve_step": {},
		"run_resume":       {},
	}
	executionRunTools = map[string]struct{}{
		"run_propose_step_result": {},
	}
	anyRoleRunTools = map[string]struct{}{
		"run_read":   {},
		"run_active": {},
		"run_note":   {},
	}
	codingProfileTools = map[string]struct{}{
		"exec":        {},
		"exec_bg":     {},
		"exec_status": {},
		"exec_stop":   {},
		"fs_read":     {},
		"fs_list":     {},
		"fs_write":    {},
		"fs_edit":     {},
		"fs_mkdir":    {},
		"fs_delete":   {},
		"fs_stat":     {},
	}
	browserProfileTools = map[string]struct{}{
		"browser_navigate":   {},
		"browser_action":     {},
		"browser_screenshot": {},
	}
	knowledgeProfileTools = map[string]struct{}{
		"search_knowledge":            {},
		"search_learnings":            {},
		"rag_retrieve":                {},
		"graph_traverse":              {},
		"graph_query":                 {},
		"save_knowledge":              {},
		"save_learning":               {},
		"create_skill":                {},
		"list_skills":                 {},
		"import_skill":                {},
		"learning_stats":              {},
		"learning_cleanup":            {},
		"librarian_pending_inquiries": {},
		"librarian_dismiss_inquiry":   {},
	}
	supervisorProfileTools = map[string]struct{}{
		"run_read":   {},
		"run_active": {},
		"run_note":   {},
	}
)

type snapshotCacheKey struct{}

type snapshotCache struct {
	mu      sync.Mutex
	entries map[string]*snapshotCacheEntry
}

type snapshotCacheEntry struct {
	once sync.Once
	snap *RunSnapshot
	err  error
}

// ToolProfileGuard returns a middleware that narrows execution tools according
// to the active step's ToolProfile for workflow/background sessions.
func ToolProfileGuard(store RunLedgerStore) toolchain.Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if store == nil {
				return next(ctx, params)
			}

			runID := runIDFromSessionContext(ctx)
			if runID == "" {
				return next(ctx, params)
			}

			snap, err := getSnapshotForTurn(ctx, store, runID)
			if err != nil {
				return next(ctx, params)
			}
			step := snap.FindStep(snap.CurrentStepID)
			if step == nil || len(step.ToolProfile) == 0 {
				return next(ctx, params)
			}

			if toolAllowedForProfiles(ctx, tool.Name, step.ToolProfile) {
				return next(ctx, params)
			}

			return nil, fmt.Errorf("tool %q is not allowed for active tool profile %v", tool.Name, step.ToolProfile)
		}
	}
}

// WithSnapshotCache seeds a per-turn snapshot cache into the context.
func WithSnapshotCache(ctx context.Context) context.Context {
	if snapshotCacheFromContext(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, snapshotCacheKey{}, &snapshotCache{
		entries: make(map[string]*snapshotCacheEntry),
	})
}

func snapshotCacheFromContext(ctx context.Context) *snapshotCache {
	cache, _ := ctx.Value(snapshotCacheKey{}).(*snapshotCache)
	return cache
}

func (c *snapshotCache) load(
	runID string,
	loadFn func() (*RunSnapshot, error),
) (*RunSnapshot, error) {
	c.mu.Lock()
	entry := c.entries[runID]
	if entry == nil {
		entry = &snapshotCacheEntry{}
		c.entries[runID] = entry
	}
	c.mu.Unlock()

	entry.once.Do(func() {
		entry.snap, entry.err = loadFn()
	})
	return entry.snap, entry.err
}

func getSnapshotForTurn(
	ctx context.Context,
	store RunLedgerStore,
	runID string,
) (*RunSnapshot, error) {
	cache := snapshotCacheFromContext(ctx)
	if cache == nil {
		return store.GetRunSnapshot(ctx, runID)
	}
	return cache.load(runID, func() (*RunSnapshot, error) {
		return store.GetRunSnapshot(ctx, runID)
	})
}

func runIDFromSessionContext(ctx context.Context) string {
	rc := session.RunContextFromContext(ctx)
	if rc == nil {
		return ""
	}
	return rc.RunID
}

func toolAllowedForProfiles(ctx context.Context, toolName string, profiles []string) bool {
	if _, ok := anyRoleRunTools[toolName]; ok {
		return true
	}

	agentName := toolchain.AgentNameFromContext(ctx)
	if isOrchestratorAgentName(agentName) {
		_, ok := orchestratorOnlyRunTools[toolName]
		return ok
	}
	if agentName != "" {
		if _, ok := executionRunTools[toolName]; ok {
			return true
		}
	}

	for _, profile := range profiles {
		switch ToolProfile(profile) {
		case ToolProfileCoding:
			if _, ok := codingProfileTools[toolName]; ok {
				return true
			}
		case ToolProfileBrowser:
			if _, ok := browserProfileTools[toolName]; ok {
				return true
			}
		case ToolProfileKnowledge:
			if _, ok := knowledgeProfileTools[toolName]; ok {
				return true
			}
		case ToolProfileSupervisor:
			if _, ok := supervisorProfileTools[toolName]; ok {
				return true
			}
		}
	}
	return false
}
