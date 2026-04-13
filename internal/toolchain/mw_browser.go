package toolchain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/tools/browser"
)

// WithBrowserRecovery returns a middleware that provides panic recovery and
// auto-reconnect for browser tools. It only applies to tools whose name
// starts with "browser_"; other tools pass through unchanged.
func WithBrowserRecovery(sm *browser.SessionManager) Middleware {
	return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		if !strings.HasPrefix(tool.Name, "browser_") {
			return next
		}
		return func(ctx context.Context, params map[string]interface{}) (result interface{}, retErr error) {
			defer func() {
				if r := recover(); r != nil {
					logging.App().Errorw("browser tool panic recovered", "tool", tool.Name, "panic", r)
					retErr = fmt.Errorf("%w: %v", browser.ErrBrowserPanic, r)
				}
			}()

			result, retErr = next(ctx, params)
			if retErr != nil && errors.Is(retErr, browser.ErrBrowserPanic) {
				// Connection likely dead — close and retry once.
				logging.App().Warnw("browser panic detected, closing session and retrying", "tool", tool.Name, "error", retErr)
				_ = sm.Close()
				result, retErr = next(ctx, params)
			}

			// CDP target error recovery — navigate only (no side-effect tools).
			if retErr != nil && tool.Name == "browser_navigate" {
				if strings.Contains(retErr.Error(), "Inspected target navigated or closed") {
					logging.App().Warnw("browser navigate CDP target error, resetting session and retrying",
						"tool", tool.Name, "error", retErr)
					_ = sm.Close()
					result, retErr = next(ctx, params)
				}
			}
			return
		}
	}
}
