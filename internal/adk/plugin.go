// Package adk — ADK Plugin Integration Spike
//
// SPIKE ONLY — not wired to production runtime.
// These plugin constructors demonstrate ADK callback mapping feasibility.
// They are NOT called from app wiring (internal/app/wiring.go).
// Production integration requires EventBus bridge design (2nd phase).
//
// # Parity Gap Analysis: ADK plugin callbacks vs Lango toolchain middleware
//
// ## Summary
//
// ADK v0.6.0 provides agent-level callback hooks via plugin.Config.
// Lango provides per-tool middleware via internal/toolchain (Chain/ChainAll)
// with a priority-based hook registry.
//
// ## Key Structural Differences
//
// 1. SCOPE: ADK callbacks are agent-level (fire for ALL tools on the agent).
//    Lango middleware supports per-tool application — Chain() wraps a single
//    tool, and middlewares like WithBrowserRecovery skip non-matching tools
//    by inspecting tool.Name inside the middleware itself.
//    → This is the CRITICAL gap. ADK has no way to selectively apply
//      callbacks to specific tools without manual name-checking inside the
//      callback body.
//
// 2. PRIORITY/ORDERING: ADK callbacks execute in plugin registration order.
//    There is no priority field — only insertion order matters.
//    Lango HookRegistry supports numeric Priority (lower = earlier) with
//    automatic sorting.
//    → Cannot reorder ADK callbacks without changing plugin registration order.
//
// 3. BLOCKING: ADK BeforeToolCallback can block execution by returning a
//    non-nil result (the tool call is skipped, and the returned map is used
//    as the tool result). Returning a non-nil error also blocks.
//    Lango PreToolHook uses explicit PreHookAction (Continue/Block/Modify)
//    with a BlockReason string.
//    → Functional parity for blocking. ADK's mechanism is more implicit
//      (return non-nil = block), Lango's is more explicit (enum action).
//
// 4. PARAM MODIFICATION: ADK BeforeToolCallback can modify args in-place
//    and return (nil, nil) to proceed with modified args.
//    Lango PreToolHook returns Modify action with ModifiedParams map.
//    → Functional parity, different ergonomics.
//
// 5. RESULT MODIFICATION: ADK AfterToolCallback receives (args, result, err)
//    and can return a replacement result map. Non-nil return replaces the
//    actual tool result.
//    Lango PostToolHook receives (result, error) but its return error is
//    logged, not propagated — it cannot modify the result.
//    → ADK AfterToolCallback is MORE powerful here (can rewrite results).
//
// 6. ERROR HANDLING: ADK provides OnToolErrorCallback specifically for error
//    cases. Lango has no separate error callback — errors flow through the
//    middleware chain naturally.
//    → ADK's dedicated error callback is a nice separation of concerns.
//
// ## Per-Middleware Migration Assessment
//
// | Lango Middleware/Hook             | Can Move to ADK Plugin? | Notes                                                    |
// |-----------------------------------|-------------------------|----------------------------------------------------------|
// | SecurityFilterHook (pre, P:10)    | PARTIAL                 | BeforeToolCallback can block. But must manually check     |
// |                                   |                         | tool names — no per-tool scoping. Security-critical code  |
// |                                   |                         | should NOT rely on implicit "return non-nil to block."    |
// |                                   |                         | RECOMMENDATION: Keep in toolchain.                       |
// | AgentAccessControlHook (pre, P:20)| PARTIAL                 | BeforeAgentCallback could check agent-level ACL, but     |
// |                                   |                         | tool-level ACL needs BeforeToolCallback. No agent name    |
// |                                   |                         | in tool.Context easily available.                         |
// |                                   |                         | RECOMMENDATION: Keep in toolchain.                       |
// | EventBusHook (pre+post, P:50)     | YES                     | Good candidate. OnEventCallback can observe all events.   |
// |                                   |                         | BeforeTool/AfterTool can measure duration.                |
// |                                   |                         | Agent-level scope is fine for observability.              |
// | KnowledgeSaveHook (post, P:100)   | PARTIAL                 | AfterToolCallback receives result. But per-tool filtering |
// |                                   |                         | (SaveableTools set) must be done inside the callback.     |
// |                                   |                         | RECOMMENDATION: Could move, but loses opt-in clarity.    |
// | WithLearning (middleware)         | YES                     | AfterToolCallback receives all needed data.               |
// |                                   |                         | Good candidate for migration.                            |
// | WithApproval (middleware)         | NO                      | Needs per-tool scoping (NeedsApproval checks policy +    |
// |                                   |                         | tool safety level). Needs external approval provider,     |
// |                                   |                         | grant store, spending limiter. Too much Lango-specific    |
// |                                   |                         | logic.                                                   |
// | WithBrowserRecovery (middleware)  | NO                      | Per-tool (browser_* prefix only). Needs SessionManager   |
// |                                   |                         | reference. Panic recovery with defer/recover.             |
// |                                   |                         | ADK callbacks don't support panic recovery.               |
// | WithOutputManager (middleware)    | PARTIAL                 | AfterToolCallback could post-process output. But tightly  |
// |                                   |                         | coupled to Lango's token budget and compression system.   |
// |                                   |                         | RECOMMENDATION: Keep in toolchain.                       |
// | WithTruncate (middleware)         | YES                     | Simple output transformation. AfterToolCallback can do    |
// |                                   |                         | this. But it's so simple it doesn't benefit from the move.|
// | WithHooks (middleware)            | N/A                     | This IS the hook registry adapter. Not applicable.        |
//
// ## Conclusion
//
// Only pure observability hooks (EventBusHook, WithLearning) are clean migration
// candidates. Security, approval, per-tool middleware, and Lango-specific logic
// must remain in toolchain. The recommended strategy is:
//   - Use ADK plugins for NEW cross-cutting concerns (logging, tracing, metrics)
//   - Keep existing toolchain for Lango-specific per-tool middleware
//   - Do NOT migrate security-critical hooks to ADK plugins

package adk

import (
	adk_agent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"

	"github.com/langoai/lango/internal/logging"
)

// NewEventLoggingPlugin creates an ADK plugin that logs session events for
// observability. This is a spike implementation demonstrating OnEventCallback
// integration.
//
// The plugin logs:
//   - Event author and type
//   - Delegation transfers (agent-to-agent)
//   - Function calls (tool invocations)
//   - Text responses
//
// This demonstrates a cross-cutting concern that maps cleanly to ADK's
// agent-level plugin model — no per-tool scoping needed.
//
// Spike implementation. Not called from app wiring.
func NewEventLoggingPlugin() (*plugin.Plugin, error) {
	return plugin.New(plugin.Config{
		Name: "lango-event-logger",

		OnEventCallback: func(_ adk_agent.InvocationContext, evt *session.Event) (*session.Event, error) {
			log := logging.Agent()

			if evt.Actions.TransferToAgent != "" {
				log.Debugw("plugin: agent delegation observed",
					"from", evt.Author,
					"to", evt.Actions.TransferToAgent)
			}

			if evt.Content != nil {
				for _, p := range evt.Content.Parts {
					if p.FunctionCall != nil {
						log.Debugw("plugin: tool call observed",
							"author", evt.Author,
							"tool", p.FunctionCall.Name)
					}
					if p.Text != "" {
						log.Debugw("plugin: text response observed",
							"author", evt.Author,
							"partial", evt.Partial,
							"text_len", len(p.Text))
					}
				}
			}

			// Return the event unmodified — pure observation.
			return evt, nil
		},
	})
}

// --- Spike: BeforeToolCallback → SecurityFilterHook mapping ---
//
// ADK's BeforeToolCallback signature:
//   func(ctx tool.Context, t tool.Tool, args map[string]any) (map[string]any, error)
//
// Mapping to SecurityFilterHook:
//   - tool.Tool.Name() provides the tool name for blocklist checking
//   - args["command"] provides the command string for pattern matching
//   - Return non-nil map to BLOCK (the map becomes the tool result, skipping execution)
//   - Return (nil, nil) to ALLOW
//
// Example (NOT wired — spike only):
//
//   func securityFilterAsBeforeTool(blocked []string) llmagent.BeforeToolCallback {
//       return func(ctx tool.Context, t tool.Tool, args map[string]any) (map[string]any, error) {
//           cmd, _ := args["command"].(string)
//           for _, pattern := range blocked {
//               if strings.Contains(strings.ToLower(cmd), strings.ToLower(pattern)) {
//                   // Return non-nil to block execution.
//                   return map[string]any{
//                       "error": "blocked by security policy: " + pattern,
//                   }, nil
//               }
//           }
//           return nil, nil // allow
//       }
//   }
//
// GAPS vs Lango SecurityFilterHook:
//   1. No per-tool scoping — this fires for ALL tools, not just exec-like tools
//   2. No BlockReason propagation — must encode reason in the result map
//   3. No Priority field — ordering depends on plugin registration order
//   4. Blocked tools list checking requires manual tool.Name() matching
//      instead of the HookRegistry's structured approach

// --- Spike: AfterToolCallback → WithLearning middleware mapping ---
//
// ADK's AfterToolCallback signature:
//   func(ctx tool.Context, t tool.Tool, args, result map[string]any, err error) (map[string]any, error)
//
// Mapping to WithLearning:
//   - tool.Tool.Name() → toolName parameter for observer
//   - args → params
//   - result → observation result
//   - err → tool error
//   - Session key: not directly available in tool.Context (would need extraction
//     from ctx or the session service — this is a gap)
//
// Example (NOT wired — spike only):
//
//   func learningAsAfterTool(observer learning.ToolResultObserver) llmagent.AfterToolCallback {
//       return func(ctx tool.Context, t tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
//           // GAP: tool.Context does not expose session key directly.
//           // Would need to extract from context or use agent state.
//           sessionKey := "" // ← not available without Lango context bridge
//           observer.OnToolResult(ctx.(context.Context), sessionKey, t.Name(), toInterfaceMap(args), toInterface(result), err)
//           return nil, nil // pass through unmodified
//       }
//   }
//
// GAPS vs Lango WithLearning:
//   1. Session key not available in tool.Context — requires context bridge
//   2. Type conversion needed: map[string]any ↔ map[string]interface{}
//   3. tool.Context is an interface, not context.Context — may need type assertion
//   4. Result is map[string]any, not interface{} — constrains result types

// NewBeforeToolLoggingPlugin creates a spike plugin that logs tool invocations
// via BeforeToolCallback. Demonstrates the callback firing for ALL tools
// (agent-level scope) and shows how blocking would work.
//
// Spike implementation. Not called from app wiring.
func NewBeforeToolLoggingPlugin() (*plugin.Plugin, error) {
	return plugin.New(plugin.Config{
		Name: "lango-tool-logger",

		BeforeToolCallback: func(_ tool.Context, t tool.Tool, args map[string]any) (map[string]any, error) {
			log := logging.Agent()
			log.Debugw("plugin: before tool",
				"tool", t.Name(),
				"arg_count", len(args))
			// Return (nil, nil) to allow execution to proceed.
			return nil, nil
		},

		AfterToolCallback: func(_ tool.Context, t tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
			log := logging.Agent()
			if err != nil {
				log.Warnw("plugin: tool error",
					"tool", t.Name(),
					"error", err)
			} else {
				log.Debugw("plugin: after tool",
					"tool", t.Name(),
					"result_keys", mapKeys(result))
			}
			// Return (nil, nil) to pass through the original result.
			return nil, nil
		},

		OnToolErrorCallback: func(_ tool.Context, t tool.Tool, args map[string]any, err error) (map[string]any, error) {
			log := logging.Agent()
			log.Warnw("plugin: tool error callback",
				"tool", t.Name(),
				"error", err)
			// Return (nil, nil) to let the error propagate normally.
			return nil, nil
		},
	})
}

// mapKeys returns the keys of a map for logging purposes.
func mapKeys(m map[string]any) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Ensure callback type compatibility at compile time.
var (
	_ llmagent.BeforeToolCallback  = func(tool.Context, tool.Tool, map[string]any) (map[string]any, error) { return nil, nil }
	_ llmagent.AfterToolCallback   = func(tool.Context, tool.Tool, map[string]any, map[string]any, error) (map[string]any, error) { return nil, nil }
	_ llmagent.OnToolErrorCallback = func(tool.Context, tool.Tool, map[string]any, error) (map[string]any, error) { return nil, nil }
)
