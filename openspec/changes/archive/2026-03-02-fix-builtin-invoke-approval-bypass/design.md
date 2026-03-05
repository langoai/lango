## Context

The v0.3.0 multi-agent architecture enforces a strict security boundary: the orchestrator has no tools and must delegate all work to specialized sub-agents (operator, vault, navigator, etc.). Each sub-agent only has access to tools matching its role, and dangerous tools (payment, crypto, secrets) go through the ADK approval middleware chain.

Post-v0.3.0, the `toolcatalog` system introduced `builtin_invoke`, a meta-tool that can proxy-execute any registered tool. This was wired as a "universal tool" given to the orchestrator, effectively bypassing both the sub-agent role isolation and the approval middleware. The LLM could invoke `builtin_invoke(tool_name="payment_send", ...)` directly from the orchestrator context, skipping vault's approval chain entirely.

## Goals / Non-Goals

**Goals:**
- Restore the v0.3.0 security invariant: orchestrator delegates, never executes tools directly
- Block dangerous tools from being proxy-executed via `builtin_invoke`, even in non-orchestrator contexts
- Keep `builtin_invoke` functional for safe tools (e.g., `builtin_invoke("browser_navigate")` still works)
- Minimize code churn — surgical fixes at the right abstraction layer

**Non-Goals:**
- Removing the `UniversalTools` field from `Config` (may be useful for single-agent or future use)
- Redesigning the tool catalog system
- Adding a full approval middleware to the dispatcher (the dispatcher is a convenience, not a security boundary)

## Decisions

### Decision 1: Block dangerous tools at the dispatcher level

**Choice**: Add a safety level check in `builtin_invoke`'s handler before executing the tool. Tools with `SafetyLevel >= Dangerous` return an error directing the LLM to delegate.

**Why not alternatives**:
- *Remove `builtin_invoke` entirely*: Too aggressive — it's useful for safe tools in single-agent mode.
- *Wire full approval middleware into dispatcher*: Over-engineering — the correct path is sub-agent delegation, not replicating the approval chain in a second location.

### Decision 2: Stop passing universal tools to the orchestrator in multi-agent mode

**Choice**: Remove the `universalTools` construction and assignment in `wiring.go` for multi-agent mode. The `Config.UniversalTools` field is left in the struct (not deleted).

**Why**: The orchestrator's entire purpose is delegation. Giving it tools undermines this. Keeping the field allows future single-agent use cases.

### Decision 3: Simplify `buildOrchestratorInstruction` signature

**Choice**: Remove the `hasUniversalTools bool` parameter. The orchestrator always emits "You do NOT have tools" in its instruction.

**Why**: With universal tools removed from the orchestrator, the conditional branch is dead code. Removing it prevents accidental re-enablement.

## Risks / Trade-offs

- [Risk] Single-agent mode may want `builtin_invoke` for dangerous tools → Mitigation: This fix only blocks at the dispatcher; single-agent tools go through normal ADK middleware, not the dispatcher.
- [Risk] LLM may still hallucinate direct tool calls → Mitigation: The orchestrator prompt explicitly says "You do not have direct access to tools" and the orchestrator has no tools in its ADK config.
- [Trade-off] `builtin_invoke` is now less powerful (can't proxy dangerous tools) → Acceptable: dangerous tools should always go through proper approval chains.
