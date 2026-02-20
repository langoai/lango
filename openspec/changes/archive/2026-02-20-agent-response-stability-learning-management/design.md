## Context

Agent responses on Telegram appear to hang indefinitely. Investigation reveals two root causes:

1. **Hallucination retry is silent**: `RunAndCollect` in `internal/adk/agent.go` retries when an invalid sub-agent name is hallucinated, but logs nothing about success or failure of the retry — making it impossible to distinguish slow retries from actual hangs.
2. **Permanent learning rate limit**: `knowledge.Store` enforces per-session limits (`maxLearnings=10`, `maxKnowledge=20`) using in-memory counters. Once hit, the session permanently loses the ability to learn. There is no way for users to manage or reclaim capacity.

The knowledge store holds a `sync.Mutex` and two `map[string]int` counters solely for this rate limiting. The config exposes `MaxLearnings` and `MaxKnowledge` fields, and the TUI settings form has corresponding input fields.

## Goals / Non-Goals

**Goals:**
- Remove all per-session rate limiting from the knowledge store
- Provide agent-accessible tools for learning data lifecycle management (stats, cleanup)
- Add comprehensive timing and result logging to the agent execution pipeline
- Make it possible to distinguish slow responses from actual hangs via logs

**Non-Goals:**
- Global (cross-session) rate limiting or quota systems
- Automatic learning data compaction or eviction policies
- Changes to the embedding or graph learning subsystems
- UI for learning data management (agent tools are the interface)

## Decisions

### Decision 1: Remove rate limits entirely (not raise them)

**Choice**: Remove `maxKnowledge`, `maxLearnings` fields, `reserveXxxSlot()` functions, and the mutex from `Store`.

**Alternatives considered**:
- **Raise limits**: Still arbitrary; doesn't solve the permanent-loss problem for very long sessions.
- **Sliding window**: Adds complexity; real issue is that users have no visibility into or control over accumulation.

**Rationale**: The rate limit was a safety net against runaway accumulation, but the real solution is user-driven cleanup via agent tools. Removing the limit simplifies the Store and eliminates a class of silent failures.

### Decision 2: Agent tools for learning management (not CLI commands)

**Choice**: Add `learning_stats` and `learning_cleanup` as agent tools in `buildMetaTools`.

**Alternatives considered**:
- **CLI commands**: Would require a separate `lango learning` subcommand tree; less discoverable for chat users.
- **TUI screen**: Too heavyweight for what is essentially a query + delete operation.

**Rationale**: Users interact with learning data through the agent. Making stats/cleanup available as agent tools means users can say "how many learnings do I have?" or "clean up old learnings" naturally in conversation.

### Decision 3: Observability at two layers (ADK + App)

**Choice**: Add timing logs both in `adk.Agent.RunAndCollect` (low-level) and `app.App.runAgent` (high-level).

**Rationale**: ADK layer captures hallucination retry details (retry count, per-retry timing). App layer captures end-to-end request lifecycle (timeout approaching, total elapsed, response size). Both are needed for effective debugging.

### Decision 4: 80% timeout early warning via `time.AfterFunc`

**Choice**: Use `time.AfterFunc(timeout * 0.8, ...)` in `runAgent` to emit a warning log when approaching timeout.

**Alternatives considered**:
- **Context with deadline + timer goroutine**: More complex, same effect.
- **Middleware-level timeout logging**: Would require refactoring the channel handler pipeline.

**Rationale**: `time.AfterFunc` is lightweight, non-blocking, and requires minimal code change. The timer is stopped on completion to prevent false warnings.

## Risks / Trade-offs

- **[Unbounded learning growth]** → Mitigated by `learning_cleanup` tool. Users can set up periodic cleanup via cron jobs or manual agent commands.
- **[Breaking change to NewStore signature]** → All callers (tests included) must update. Contained to internal packages; no external API.
- **[Config field removal]** → Viper silently ignores unknown fields, so existing config files with `maxLearnings`/`maxKnowledge` will not cause errors.
- **[Log volume increase]** → New logs are Debug/Info level for normal operations; Warn/Error only for failures and approaching timeouts.
