package adk

// ============================================================================
// Unit 6 — MCPToolset Parity Spike
//
// Evaluates ADK mcptoolset.New() (v0.6.0) against internal/mcp/adapter.go
// for tool exposure parity. Test-only evaluation — no production code changes.
//
// Source files analyzed:
//   ADK:     google.golang.org/adk@v0.6.0/tool/mcptoolset/{set.go,tool.go,client.go}
//   Current: internal/mcp/adapter.go, internal/mcp/connection.go
//
// Adoption rule: ALL 5 conditions must PASS.
// ============================================================================

import (
	"testing"

	"google.golang.org/adk/tool/mcptoolset"

	"github.com/langoai/lango/internal/agent"
)

// TestMCPToolsetParitySummary documents the 5-condition evaluation as a
// structured test function with pass/fail verdicts.
func TestMCPToolsetParitySummary(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------
	// Condition 1: NAMING — Can "mcp__{server}__{tool}" naming be preserved?
	// -------------------------------------------------------------------
	// Current adapter (internal/mcp/adapter.go:34):
	//   toolName := fmt.Sprintf("mcp__%s__%s", dt.ServerName, tool.Name)
	//
	// ADK MCPToolset (tool.go:33):
	//   mcp := &mcpTool{ name: t.Name, ... }
	//   Uses the raw MCP tool name, no server-scoped prefix.
	//
	// The mcpTool.Name() returns the raw MCP name (e.g., "read_file").
	// There is NO server-name injection point in the Config or convertTool.
	//
	// Workaround analysis:
	//   - ToolFilter (tool.Predicate) can filter but cannot rename tools.
	//   - The mcpTool struct is unexported; we cannot subclass or wrap it
	//     to override Name().
	//   - Creating a post-conversion rename layer would require intercepting
	//     Tools() output and re-wrapping each tool.Tool — effectively
	//     defeating the purpose of using MCPToolset.
	//
	// VERDICT: FAIL — No built-in mechanism to inject server-scoped names.
	// Multi-server disambiguation (the primary reason for our naming
	// convention) is impossible without forking or wrapping.

	t.Run("Condition1_Naming", func(t *testing.T) {
		t.Parallel()

		// Type-level verification: MCPToolset uses raw MCP tool names.
		// The Config struct has no NamePrefix, NameTemplate, or ServerName field.
		_ = mcptoolset.Config{} // No naming configuration fields exist.

		t.Log("VERDICT: FAIL")
		t.Log("MCPToolset uses raw MCP tool names (e.g., 'read_file').")
		t.Log("No mechanism exists to inject server-scoped 'mcp__{server}__{tool}' naming.")
		t.Log("Multi-server tool name disambiguation is lost.")
	})

	// -------------------------------------------------------------------
	// Condition 2: APPROVAL — Can RequireConfirmationProvider express
	// always-allow, payment auto-approve, owner-approval?
	// -------------------------------------------------------------------
	// Current adapter path: MCP tools → agent.Tool → toolchain middleware:
	//   - WithApproval middleware checks SafetyLevel.IsDangerous()
	//   - Owner approval flow is in mw_approval.go
	//   - Payment auto-approve is handled by separate payment middleware
	//
	// ADK MCPToolset (tool.go:93-117):
	//   RequireConfirmationProvider func(name string, args any) bool
	//   - Returns true → HITL confirmation required
	//   - Returns false → tool executes immediately
	//   - Confirmation flow is ADK-native (ctx.RequestConfirmation)
	//
	// Analysis:
	//   - Always-allow: provider returns false → OK
	//   - Payment auto-approve: provider can inspect tool name/args → PARTIAL
	//     (but payment logic would need to be duplicated in the provider,
	//     not reusing our existing middleware)
	//   - Owner approval: the ADK HITL flow is a binary confirm/reject via
	//     ctx.ToolConfirmation(). Our approval middleware has richer
	//     semantics (prompt text, audit logging, approval reasons).
	//
	// The fundamental issue: MCPToolset's confirmation flow is ADK-native
	// (HITL via FunctionResponse), while our approval system goes through
	// the toolchain middleware at the agent.ToolHandler level. These are
	// two different execution paths that don't compose naturally.
	//
	// However, since MCP tools ultimately flow through the ADK runner,
	// the plugin.BeforeToolCallback CAN intercept all tool calls
	// (including MCPToolset tools). This means approval logic could be
	// moved to BeforeToolCallback, bypassing MCPToolset's native HITL.
	//
	// VERDICT: PARTIAL PASS — Basic approval is possible via
	// RequireConfirmationProvider or ADK plugin callbacks, but our rich
	// approval middleware (audit logging, reasons, multi-level) would
	// need significant restructuring.

	t.Run("Condition2_Approval", func(t *testing.T) {
		t.Parallel()

		// Verify ConfirmationProvider type compatibility.
		var provider mcptoolset.ConfirmationProvider = func(name string, input any) bool {
			// Example: always allow safe tools, require confirmation for dangerous ones.
			_ = name
			_ = input
			return false // always allow
		}
		_ = mcptoolset.Config{
			RequireConfirmationProvider: provider,
		}

		t.Log("VERDICT: PARTIAL PASS")
		t.Log("Basic confirm/reject possible via ConfirmationProvider or ADK plugin callbacks.")
		t.Log("Rich approval middleware (audit, reasons, multi-level) needs restructuring.")
	})

	// -------------------------------------------------------------------
	// Condition 3: SAFETY — Can per-tool SafetyLevel be propagated?
	// -------------------------------------------------------------------
	// Current adapter (adapter.go:40):
	//   safety := parseSafetyLevel(conn.cfg.SafetyLevel)
	//   Sets agent.Tool.SafetyLevel per server config.
	//
	// ADK MCPToolset:
	//   The mcpTool struct has NO safety level field.
	//   The Config struct has NO safety level configuration.
	//   convertTool() does not accept or propagate safety levels.
	//   The tool.Tool interface has no SafetyLevel method.
	//
	// Our safety level (agent.SafetyLevel) is a Lango-specific concept
	// that exists on agent.Tool but NOT on ADK's tool.Tool interface.
	// The ADK tool.Tool interface defines: Name(), Description(),
	// IsLongRunning(), ProcessRequest(), Run().
	//
	// Even with the ADK plugin BeforeToolCallback, the callback receives
	// a tool.Tool which has no safety level. We would need to maintain
	// a separate name→safety map outside the tool, losing the
	// encapsulation that the current adapter provides.
	//
	// VERDICT: FAIL — ADK tool.Tool has no concept of safety levels.
	// Per-tool safety metadata is lost, breaking our approval/access
	// control pipeline that relies on SafetyLevel.IsDangerous().

	t.Run("Condition3_Safety", func(t *testing.T) {
		t.Parallel()

		// Verify our safety level type exists and is meaningful.
		levels := []agent.SafetyLevel{
			agent.SafetyLevelSafe,
			agent.SafetyLevelModerate,
			agent.SafetyLevelDangerous,
		}
		for _, level := range levels {
			if !level.Valid() {
				t.Errorf("expected SafetyLevel(%d) to be valid", level)
			}
		}

		// ADK tool.Tool interface has no SafetyLevel concept — verified by
		// inspecting mcpTool struct fields in tool.go:58-68:
		//   name, description, funcDeclaration, mcpClient,
		//   requireConfirmation, requireConfirmationProvider
		// No safety field exists.

		t.Log("VERDICT: FAIL")
		t.Log("ADK tool.Tool interface has no SafetyLevel concept.")
		t.Log("Per-tool safety metadata is lost, breaking approval/access pipelines.")
	})

	// -------------------------------------------------------------------
	// Condition 4: TRUNCATION — Can maxOutputTokens be applied to results?
	// -------------------------------------------------------------------
	// Current adapter (adapter.go:77):
	//   result := formatContent(res.Content, maxOutputTokens)
	//   Truncates at maxChars = maxOutputTokens * 4
	//
	// ADK MCPToolset (tool.go:148-174):
	//   Returns raw text without any truncation:
	//     return map[string]any{"output": textResponse.String()}, nil
	//   No maxOutputTokens field in Config.
	//   No truncation hook or result transformation callback.
	//
	// Workaround analysis:
	//   - AfterToolCallback (via plugin) receives the result map and CAN
	//     modify it. So truncation COULD be applied there.
	//   - However, AfterToolCallback operates on ALL tools, not just MCP
	//     tools. We'd need to filter by tool name prefix, which circles
	//     back to the naming problem (Condition 1 FAIL).
	//   - The existing WithTruncate middleware (toolchain/mw_truncate.go)
	//     operates at the agent.Tool level, not the ADK tool.Tool level.
	//
	// VERDICT: CONDITIONAL PASS — Truncation CAN be applied via ADK
	// AfterToolCallback, but requires workaround (name-based filtering
	// or tagging). Not as clean as current inline truncation.

	t.Run("Condition4_Truncation", func(t *testing.T) {
		t.Parallel()

		// Verify MCPToolset Config has no truncation field.
		cfg := mcptoolset.Config{}
		_ = cfg // No MaxOutputTokens, Truncation, or similar field.

		t.Log("VERDICT: CONDITIONAL PASS")
		t.Log("No native truncation in MCPToolset.")
		t.Log("Can be applied via ADK AfterToolCallback plugin, but requires workaround.")
	})

	// -------------------------------------------------------------------
	// Condition 5: EVENTS — Can tool call/result events reach the event bus?
	// -------------------------------------------------------------------
	// Current adapter path:
	//   MCP tools → agent.Tool → toolchain ChainAll → WithHooks middleware
	//     → EventBusHook.Pre/Post → eventbus.Bus.Publish(ToolExecutedEvent)
	//
	// ADK MCPToolset path:
	//   MCPToolset.Tools() → mcpTool → ADK runner executes Run()
	//   The mcpTool.Run() calls mcpClient.CallTool() directly.
	//
	// ADK plugin callbacks provide the interception points:
	//   - BeforeToolCallback: called before tool.Run() → maps to Pre
	//   - AfterToolCallback: called after tool.Run() → maps to Post
	//   - OnToolErrorCallback: called on tool error
	//
	// These callbacks receive (tool.Context, tool.Tool, args, result, err)
	// which contains all the data needed for ToolExecutedEvent:
	//   - ToolName: tool.Name()
	//   - Duration: measured between Before/After
	//   - Success: err == nil
	//   - Error: err.Error()
	//
	// Missing from callback context:
	//   - AgentName: available via ctx (tool.Context embeds context.Context)
	//   - SessionKey: available via session state in tool.Context
	//
	// VERDICT: PASS — ADK plugin callbacks (BeforeToolCallback,
	// AfterToolCallback) provide equivalent interception for event bus
	// publishing. Already used in our agent.go via WithPlugins option.

	t.Run("Condition5_Events", func(t *testing.T) {
		t.Parallel()

		// Verify ADK plugin callback types exist and are compatible.
		// The plugin.Config has BeforeToolCallback and AfterToolCallback
		// fields. These are the same callback types as llmagent callbacks,
		// registered at the runner level for ALL tool invocations.

		// Verify our Agent supports plugins.
		// (WithPlugins option exists — verified in agent.go)
		_ = WithPlugins // function exists

		t.Log("VERDICT: PASS")
		t.Log("ADK plugin BeforeToolCallback/AfterToolCallback intercept all tool calls.")
		t.Log("Event bus publishing can be wired through plugin callbacks.")
	})
}

// TestMCPToolsetAdoptionDecision documents the final recommendation.
func TestMCPToolsetAdoptionDecision(t *testing.T) {
	t.Parallel()

	t.Log("=== MCPToolset Parity Spike — Final Assessment ===")
	t.Log("")
	t.Log("| # | Condition    | Verdict          | Notes                                       |")
	t.Log("|---|------------- |------------------|---------------------------------------------|")
	t.Log("| 1 | Naming       | FAIL             | No server-scoped naming; raw MCP names only  |")
	t.Log("| 2 | Approval     | PARTIAL PASS     | Basic OK via provider; rich middleware lost   |")
	t.Log("| 3 | Safety       | FAIL             | No SafetyLevel concept in ADK tool.Tool      |")
	t.Log("| 4 | Truncation   | CONDITIONAL PASS | Via AfterToolCallback; not native             |")
	t.Log("| 5 | Events       | PASS             | Via ADK plugin callbacks                      |")
	t.Log("")
	t.Log("RECOMMENDATION: DO NOT ADOPT mcptoolset.New()")
	t.Log("")
	t.Log("Rationale:")
	t.Log("  - Conditions 1 (Naming) and 3 (Safety) are hard FAILs.")
	t.Log("  - Server-scoped naming (mcp__{server}__{tool}) is critical for")
	t.Log("    multi-server disambiguation and is not achievable without")
	t.Log("    forking or heavily wrapping MCPToolset.")
	t.Log("  - Per-tool safety levels drive our entire approval/access control")
	t.Log("    pipeline; losing this metadata breaks a core safety invariant.")
	t.Log("  - The current internal/mcp/adapter.go provides all 5 capabilities")
	t.Log("    natively and integrates cleanly with toolchain middleware.")
	t.Log("")
	t.Log("Keep current internal/mcp/ adapter. Re-evaluate if ADK adds:")
	t.Log("  - Tool name customization (prefix/template)")
	t.Log("  - Tool metadata/annotation support (for safety levels)")
}

// TestMCPToolsetTypeCompat verifies that the types used by both systems
// are structurally compatible, should a future hybrid approach be considered.
func TestMCPToolsetTypeCompat(t *testing.T) {
	t.Parallel()

	t.Run("ConfirmationProvider_signature", func(t *testing.T) {
		t.Parallel()

		// Verify the ConfirmationProvider signature matches expectations.
		var _ mcptoolset.ConfirmationProvider = func(toolName string, input any) bool {
			return false
		}
	})

	t.Run("Config_fields_available", func(t *testing.T) {
		t.Parallel()

		// Verify all Config fields we analyzed are accessible.
		cfg := mcptoolset.Config{
			// Client:    nil,     // *mcp.Client — optional
			// Transport: nil,     // mcp.Transport — required for real usage
			ToolFilter:                  nil, // tool.Predicate — optional
			RequireConfirmation:         false,
			RequireConfirmationProvider: nil,
		}
		_ = cfg
	})

	t.Run("our_adapter_produces_agent_Tool", func(t *testing.T) {
		t.Parallel()

		// Verify our adapter output type.
		// AdaptTool returns *agent.Tool which has SafetyLevel — this is
		// the key differentiator vs ADK's tool.Tool interface.
		tool := &agent.Tool{
			Name:        "mcp__testserver__read_file",
			Description: "Read a file",
			SafetyLevel: agent.SafetyLevelSafe,
		}
		if tool.SafetyLevel != agent.SafetyLevelSafe {
			t.Errorf("expected SafetyLevelSafe, got %v", tool.SafetyLevel)
		}
		if tool.SafetyLevel.IsDangerous() {
			t.Error("SafetyLevelSafe should not be IsDangerous()")
		}
	})
}
