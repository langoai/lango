## Context

Lango manages conversation context through a two-tier system: a DB-level sliding window (`maxHistoryTurns`, default 50) and an ADK-level hard cap (100 messages). Both use message count, not token count. When messages are trimmed, all embedded context is permanently lost. The existing 6-layer Knowledge RAG handles static knowledge but not dynamic conversation context.

The Observational Memory (OM) system introduces a middle tier between raw message history and static knowledge: compressed observation notes that preserve essential conversation context while consuming far fewer tokens.

## Goals / Non-Goals

**Goals:**
- Preserve critical conversation context (user intent, decisions, progress) beyond the message window
- Reduce token waste by compressing verbose tool results and old messages into concise observations
- Enable token-budget-based context management instead of message-count-based
- Integrate seamlessly with the existing Knowledge RAG pipeline
- Support asynchronous observation generation via Go goroutines

**Non-Goals:**
- Resource Scope (cross-channel/cross-session memory) — deferred to future phase
- Real-time streaming observation updates during generation
- Custom observer prompt configuration via UI (hardcoded prompts for now)
- Replacing the existing Knowledge/Learning system (OM complements, not replaces)

## Decisions

### D1: Package Structure — `internal/memory/`

New package `internal/memory/` with clear subcomponents:
- `observer.go` — Observer agent that generates observation notes from message history
- `reflector.go` — Reflector that condenses accumulated observations
- `token.go` — Token counting utilities
- `buffer.go` — Async observation buffer with goroutine lifecycle
- `types.go` — Shared types (Observation, Reflection, MemoryConfig)
- `store.go` — OM data persistence via Ent

**Rationale**: Isolated from existing packages. Core Developer owns `internal/memory/`, Application Developer integrates with `internal/adk/` and `internal/app/`.

### D2: Token Counting — Character-Based Approximation

Use a simple character-based approximation (1 token ≈ 4 characters for English, 1 token ≈ 2 characters for CJK) instead of tiktoken-go.

**Alternatives considered:**
- `tiktoken-go`: Accurate but adds heavy dependency, only works for OpenAI tokenizer
- `go-tokenizer`: Limited provider support
- Character approximation: Simple, zero dependencies, sufficient for threshold-based decisions

**Rationale**: OM uses token counts only for threshold comparisons ("should I observe now?"), not for precise API billing. A ±20% approximation is acceptable. Can upgrade to tiktoken-go later if precision becomes critical.

### D3: Observer Model — Configurable, Defaults to Primary Agent Model

The Observer and Reflector use LLM calls to generate compressed notes. The model is configurable via `observationalMemory.provider` and `observationalMemory.model` in config. Defaults to the primary agent's provider/model.

**Rationale**: Users can point Observer at a cheaper/faster model (Gemini Flash, Ollama local) to reduce cost, while defaulting to the main model for simplicity.

### D4: Observation Trigger — Token Threshold on Message History

Observation is triggered when the total token count of un-observed messages exceeds `messageTokenThreshold` (default: 1000 tokens). The Observer processes these messages and produces a compressed observation note.

**Alternatives considered:**
- Message count trigger: Current approach, ignores token weight
- Time-based trigger: Inconsistent — idle sessions waste resources
- Per-message trigger: Too frequent, high LLM cost

**Rationale**: Token threshold balances cost (fewer LLM calls) with freshness (doesn't wait too long). 1000 tokens ≈ 5-10 typical user/assistant exchanges.

### D5: Reflection Trigger — Token Threshold on Observations

When accumulated observation tokens exceed `observationTokenThreshold` (default: 2000 tokens), the Reflector condenses all observations into a single reflection note. Old observations are then replaced.

**Rationale**: Prevents observation accumulation from becoming its own token problem. Two-tier compression provides progressive summarization.

### D6: Context Assembly Order

The augmented system prompt follows this order:
```
1. Base System Prompt
2. Knowledge RAG (existing 6 layers)
3. Reflections (if any)
4. Observations (if any, most recent first)
5. Recent Messages (within token budget)
```

**Rationale**: Reflections are the most compressed (oldest context), observations are middle-aged context, recent messages are exact. This ordering mirrors temporal distance from the current moment.

### D7: Async Buffering — Goroutine with Channel

Observer runs as a background goroutine triggered via a channel signal. When a new message is appended and the token threshold is exceeded, a signal is sent. The goroutine processes observations asynchronously, storing results to the DB. The next LLM call picks up stored observations.

```
AppendMessage → check token threshold → signal triggerCh
                                              ↓
                              Observer goroutine (background)
                                              ↓
                              Store observation to DB
                                              ↓
                              Next GenerateContent picks up
```

**Rationale**: Go goroutines are lightweight and the existing App already manages goroutine lifecycles via `sync.WaitGroup`. Non-blocking observation means zero latency impact on user responses.

### D8: Ent Schema — Observation and Reflection Entities

Two new Ent schemas with session edges:
- `Observation`: session_key, content, token_count, source_start_index, source_end_index, created_at
- `Reflection`: session_key, content, token_count, generation, created_at

Both reference sessions by key (string) rather than foreign key edge, matching the existing Message pattern's flexibility.

**Rationale**: Ent's auto-migration handles schema changes. String-based session reference avoids complex edge relationships while maintaining queryability.

### D9: EventsAdapter Modification — Token-Budget Dynamic Truncation

Replace the hard `if len(msgs) > 100` cap with token-budget-based truncation:
1. Calculate total token budget for messages (configurable, default: 8000 tokens)
2. Iterate from most recent message backward
3. Include messages until budget is exhausted
4. Remaining older messages are candidates for observation

**Rationale**: Preserves more messages when they're short, fewer when they contain large tool results. More intelligent than a flat count.

## Risks / Trade-offs

**[Observer quality depends on prompt engineering]** → Start with well-tested prompt templates adapted from Mastra's approach. Iterate based on real conversation quality metrics.

**[Additional LLM cost per conversation]** → Configurable model selection allows using cheaper models. Async buffering means observations batch naturally (not per-message). Default thresholds minimize calls (~1 observation per 5-10 exchanges).

**[SQLite concurrent writes from observer goroutine]** → Ent already uses connection pooling. WAL mode is enabled by default in modern SQLite. The observation write is a simple INSERT, not conflicting with message appends.

**[Token approximation inaccuracy]** → Character-based counting may over/under-estimate by 20%. For threshold triggers, this is acceptable. Worst case: observation triggers slightly early or late.

**[Observer LLM call failure]** → Graceful degradation: if observation fails, the system continues with raw message history as it does today. Failed observations are retried on next trigger. No data loss.

## Migration Plan

1. Add Ent schemas → `go generate` → auto-migration on startup
2. Add `internal/memory/` package with all components
3. Modify `ContextAwareModelAdapter` to inject observations
4. Modify `EventsAdapter` for token-budget truncation
5. Add config section and wire into `internal/app/`
6. Default: OM disabled (`enabled: false`), zero impact on existing users

**Rollback**: Set `observationalMemory.enabled: false` in config. Observation/Reflection tables remain but are ignored. No schema rollback needed.

## Open Questions

- Should observations be encrypted like session messages when SQLCipher is enabled? (Likely yes, since they contain conversation summaries)
- Optimal default token thresholds need tuning with real conversations
- Should the observer prompt be different for different languages (Korean vs English conversations)?
