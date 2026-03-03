## Context

Lango's current multi-agent system uses 7 hardcoded `AgentSpec` structs with static prefix-based tool routing and keyword-based orchestrator routing. This works but limits extensibility — adding or customizing agents requires code changes. The system also lacks tool lifecycle hooks, sub-agent context isolation, agent-scoped memory, and P2P team coordination.

The upgrade builds on existing infrastructure: `internal/orchestration/` (agent tree, tool partitioning), `internal/toolchain/` (middleware chain), `internal/eventbus/` (typed pub/sub), `internal/p2p/` (protocol, discovery, settlement), and `internal/session/` (Ent-backed session store).

## Goals / Non-Goals

**Goals:**
- Declarative agent definition via AGENT.md files with override semantics
- Dynamic tool partitioning driven by registry specs instead of hardcoded maps
- Tool execution hooks (pre/post) with priority-based ordering
- Sub-agent context isolation via child sessions
- Agent-scoped persistent memory (in-memory store)
- P2P agent pool with weighted scoring and team coordination
- CLI updates reflecting the dynamic registry

**Non-Goals:**
- LLM-based semantic routing (staying with keyword + capability matching at prompt level)
- Persistent agent memory via Ent/database (using in-memory store; Ent schema deferred)
- New payment infrastructure (reusing existing PayGate/Settlement)
- Agent marketplace or discovery protocol changes

## Decisions

### 1. AGENT.md Format (YAML frontmatter + markdown body)

Reuses the same pattern as SKILL.md: YAML frontmatter for structured metadata, markdown body for the instruction. The `splitFrontmatter` function from `skill/parser.go` is the proven pattern.

**Alternative**: JSON/TOML config files — rejected because markdown body provides better readability for agent instructions.

### 2. Three-Tier Override Semantics (User > Embedded > Builtin)

- **Builtin**: Programmatic registration (reserved for future use)
- **Embedded**: `embed.FS` containing default AGENT.md files (replaces hardcoded specs)
- **User**: `~/.lango/agents/<name>/AGENT.md` for customization

Override by name — a user-defined agent with the same name replaces the embedded default.

**Alternative**: Merge semantics (combine user + embedded) — rejected for complexity and unpredictable behavior.

### 3. Hook System as Middleware Bridge

`WithHooks(registry)` returns a `toolchain.Middleware` that integrates with the existing `Chain`/`ChainAll` infrastructure. Hooks execute in priority order (lower number = earlier execution).

**Alternative**: Separate hook execution pipeline — rejected to avoid parallel middleware systems.

### 4. Child Session "Read Parent, Write Child" Isolation

Child sessions can read parent history but write only to their own store. Results are merged back via `StructuredSummarizer` (extracts last assistant response, zero-cost) or `LLMSummarizer` (opt-in, uses LLM for summarization).

**Alternative**: Full session copy — rejected for memory/storage overhead.

### 5. In-Memory Agent Memory Store

Uses `sync.RWMutex`-protected maps instead of Ent schema. Simpler to implement and sufficient for single-process deployment. Ent schema can be added later for persistence.

**Alternative**: Immediate Ent schema — deferred to avoid migration complexity in this change.

### 6. P2P Routing via Prompt Table (Not Direct Sub-Agents)

P2P agents appear in the orchestrator's routing table but are invoked via `p2p_invoke` tool, not as direct ADK sub-agents. This is because ADK's `Agent` interface has an unexported `internal()` method that prevents external implementation.

**Alternative**: Wrapper agents that delegate to P2P — rejected for unnecessary indirection.

### 7. Weighted Agent Scoring

Trust (0.35) + Capability (0.25) + Performance (0.20) + Price (0.15) + Availability (0.05). Trust dominates because P2P interactions require reliability. Price is weighted low because quality matters more than cost for agent tasks.

## Risks / Trade-offs

- **[Registry load failure]** → Non-fatal for user stores (embedded always works). User store errors logged but don't prevent startup.
- **[Hook ordering conflicts]** → Priority-based ordering with well-separated default priorities (10, 20, 50, 100). Custom hooks should use priorities > 200.
- **[Child session memory growth]** → Sessions are short-lived (per sub-agent invocation). StructuredSummarizer keeps only the last response.
- **[In-memory agent memory loss on restart]** → Acceptable trade-off for v2. Persistence via Ent deferred to future work.
- **[P2P agent scoring gaming]** → Mitigated by trust score's dominant weight and existing reputation system.

## Migration Plan

1. All changes are additive — no existing behavior modified when new features are disabled
2. Default AGENT.md files reproduce exact behavior of current hardcoded specs
3. `PartitionTools()` preserved as backward-compatible wrapper over `PartitionToolsDynamic()`
4. New config fields have sensible defaults (hooks disabled, empty agents dir, etc.)
5. Rollback: revert the commit; no data migration needed
