package runledger

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
)

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

			snap, err := store.GetRunSnapshot(ctx, runID)
			if err != nil {
				return next(ctx, params)
			}
			step := snap.FindStep(snap.CurrentStepID)
			if step == nil || len(step.ToolProfile) == 0 {
				return next(ctx, params)
			}

			if toolAllowedForProfiles(tool.Name, step.ToolProfile) {
				return next(ctx, params)
			}

			return nil, fmt.Errorf("tool %q is not allowed for active tool profile %v", tool.Name, step.ToolProfile)
		}
	}
}

func runIDFromSessionContext(ctx context.Context) string {
	sessionKey := session.SessionKeyFromContext(ctx)
	if sessionKey == "" {
		return ""
	}
	if strings.HasPrefix(sessionKey, "bg:") {
		return strings.TrimPrefix(sessionKey, "bg:")
	}
	if strings.HasPrefix(sessionKey, "workflow:") {
		parts := strings.Split(sessionKey, ":")
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return ""
}

func toolAllowedForProfiles(toolName string, profiles []string) bool {
	// Allow run tools needed to inspect/update the active run.
	if strings.HasPrefix(toolName, "run_") {
		return true
	}
	for _, profile := range profiles {
		switch ToolProfile(profile) {
		case ToolProfileCoding:
			if strings.HasPrefix(toolName, "exec") || strings.HasPrefix(toolName, "fs_") {
				return true
			}
		case ToolProfileBrowser:
			if strings.HasPrefix(toolName, "browser_") {
				return true
			}
		case ToolProfileKnowledge:
			if strings.HasPrefix(toolName, "search_") ||
				strings.HasPrefix(toolName, "rag_") ||
				strings.HasPrefix(toolName, "graph_") ||
				strings.HasPrefix(toolName, "learning_") ||
				strings.HasPrefix(toolName, "librarian_") {
				return true
			}
		case ToolProfileSupervisor:
			if toolName == "run_read" || toolName == "run_active" || toolName == "run_note" {
				return true
			}
		}
	}
	return false
}
