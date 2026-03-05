## Context

The `feature/tui-cli-cmd-update` branch added 16K+ lines across 279 files introducing agent registry, agent memory, tool catalog, tool hooks, P2P teams, settlement, and child sessions. Three independent review passes found 22 issues grouped by severity: 3 correctness bugs, 4 code reuse problems, 5 quality/safety issues, and 4 efficiency improvements. This design covers the 16 high-impact fixes.

## Goals / Non-Goals

**Goals:**
- Fix all 3 correctness bugs (context key mismatch, event bus isolation, wrong sentinel)
- Eliminate code duplication across 5 areas (schema builder, ParseUSDC, splitFrontmatter, truncate, context keys)
- Address safety issues (fire-and-forget goroutines, data races, parameter sprawl)
- Improve efficiency in hot paths (health checks, ledger cleanup, regex elimination, index lookup)

**Non-Goals:**
- Duplicate event types between team_events and events (intentional separation)
- `MajorityResolver` naming (documented as placeholder)
- `ActiveTeams` alias (convenience method, low cost)
- `ToolExecutedEvent.Duration` always zero (future fill-in)
- Agent definition duplication (Go structs vs AGENT.md) — fallback design

## Decisions

### D1: Delegate toolchain context keys to ctxkeys (not merge packages)
**Choice**: Replace `toolchain.contextKey`/`agentNameCtxKey` with `var` aliases pointing to `ctxkeys.WithAgentName`/`ctxkeys.AgentNameFromContext`.
**Rationale**: Preserves the public API (`toolchain.AgentNameFromContext` still works) while using the single canonical context key. Moving all callers to import `ctxkeys` directly would be a larger blast radius.

### D2: Single event bus (not bus routing/forwarding)
**Choice**: Pass the global `bus` to `initP2P` instead of creating `p2pBus`.
**Rationale**: The event bus is lightweight and typed; settlement subscribes to `ToolExecutionPaidEvent` which is published by `EventBusHook` on the global bus. A separate bus breaks this subscription chain. Event namespacing (local vs P2P) is handled by event type, not bus instance.

### D3: Shared mdparse package (not utility package)
**Choice**: Create `internal/mdparse/` with only `SplitFrontmatter` rather than a generic `internal/util/`.
**Rationale**: Go style guide prohibits "util" packages. `mdparse` is specific and descriptive. Both `skill/parser.go` and `agentregistry/parser.go` use `var splitFrontmatter = mdparse.SplitFrontmatter` to minimize call-site changes.

### D4: Keep adk/agent.go truncate (import cycle avoidance)
**Choice**: Replace 5 out of 6 `truncate` copies with `toolchain.Truncate` delegation, but keep the copy in `adk/agent.go`.
**Rationale**: `adk` cannot import `toolchain` due to the dependency direction (toolchain depends on agent types). The adk copy is rune-aware and has a comment noting the canonical version.

### D5: DefaultPostPayThreshold = 0.7 (not 0.8)
**Choice**: Set the shared constant to 0.7 (matching `team/payment.go`).
**Rationale**: The team payment negotiation layer is the consumer-facing decision point. The paygate threshold was defensive but inconsistent. 0.7 aligns both layers and is the less restrictive option, which is appropriate since post-pay still requires settlement.

### D6: agentDeps struct (not builder pattern)
**Choice**: Simple struct with named fields rather than a functional options or builder pattern.
**Rationale**: The 14 parameters are all required dependencies, not optional configuration. A struct groups them without adding abstraction overhead.

## Risks / Trade-offs

- **[Risk] DefaultPostPayThreshold change from 0.8 to 0.7** → Peers that previously required prepay (score 0.7–0.8) will now qualify for post-pay. Mitigated by the fact that settlement still occurs; only timing changes.
- **[Risk] ParseUSDC var alias may confuse IDE navigation** → Mitigated by keeping the var name identical and adding a doc comment explaining the delegation.
- **[Risk] parentIndex in InMemoryChildStore adds memory overhead** → Minimal: one string slice per parent. The index eliminates O(n) full scans for `ChildrenOf()`.
- **[Risk] Parallel health checks may spike network usage** → Bounded by pool size (typically <50 agents). WaitGroup ensures completion before next tick.
