## Tasks

### Task 1: Create spike report document
- **File**: `internal/streamx/SPIKE_REPORT.md`
- **Status**: DONE
- **Description**: Analyze ADK tool execution ownership and document findings
- [x] Read `internal/adk/agent.go` — runner.Run integration, event iteration loop
- [x] Read `internal/adk/plugin.go` — BeforeToolCallback, AfterToolCallback signatures and scope
- [x] Read `internal/adk/tools.go` — AdaptTool/adaptToolWithOptions handler wrapping
- [x] Read `internal/adk/model.go` — toolCallAccumulator, streaming FunctionCall batching
- [x] Read `internal/toolchain/middleware.go` — Chain/ChainAll middleware wrapping
- [x] Read `internal/toolchain/mw_hooks.go` — WithHooks turn-level middleware
- [x] Read `internal/toolchain/hook_registry.go` — HookRegistry pre/post hook execution
- [x] Document complete flow: Model response → FunctionCall parsing → tool dispatch
- [x] Evaluate 6 hook point options (A-F) with feasibility/risk/effort
- [x] Document ADK v0.6.0 constraints on concurrent tool calls
- [x] Catalog existing reusable assets for future implementation
- [x] Provide recommendation with next steps

### Task 2: ADK runner source audit (DEFERRED)
- **Status**: BLOCKED — Go module cache inaccessible during spike
- **Description**: Read actual ADK runner source to verify sequential dispatch assumption
- [ ] Read `$GOMODCACHE/google.golang.org/adk@v0.6.0/runner/runner.go`
- [ ] Confirm whether runner has internal concurrent tool dispatch support
- [ ] Check for ToolExecutor interface or similar extension point
