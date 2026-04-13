## Context

Lango's `internal/app/` package has accumulated 76 files (17.7K LOC) through rapid feature addition. Phase B post-build wiring lacks rollback, cross-domain coupling uses two competing mechanisms (EventBus + SetXxxCallback), tool builders are locked inside app-private glue structs, and the monolithic config package is imported 196 times. Six CLI packages have zero tests. This refactoring addresses these issues in 6 independently-mergeable units.

## Goals / Non-Goals

**Goals:**
- Phase B wiring failures clean up previously registered components (rollback)
- Single coupling mechanism for cross-domain integration (EventBus)
- Domain packages own their tool builders (pilot: economy)
- Domain packages decouple from monolithic config via narrow reader interfaces (pilot: P2P)
- Eliminate duplicate interface definitions (TextGenerator → internal/llm/)
- CLI packages have baseline test coverage via shared harness

**Non-Goals:**
- Full extraction of all tool builders (only economy pilot)
- Config reader interfaces for all domains (only P2P pilot)
- Removing domain-internal observer hooks (negotiation.SetEventCallback, etc.)
- App struct field grouping into composite structs
- Automation facade (cron/background/workflow unification)

## Decisions

### 1. cleanupStack for Phase B rollback (not defer-based)
Cleanup functions accumulate in a slice during Phase B; on failure they execute in reverse order. On success the slice is discarded. This mirrors the existing bootstrap pipeline pattern in `bootstrap/pipeline.go`. Defer-based cleanup was considered but rejected because deferred functions would run unconditionally on function exit, requiring explicit success-flag checking which is more error-prone.

### 2. NeedsGraph field on ContentSavedEvent (not separate event types)
The original design used a single `ContentSavedEvent` for both embed and graph paths, which caused graph pollution on knowledge updates and learning saves. Instead of splitting into separate `EmbedRequestedEvent` and `GraphRequestedEvent`, a `NeedsGraph bool` field was added. This preserves the original SetGraphCallback semantics: only new knowledge creation and memory observations/reflections trigger graph processing. A single event type with a routing field is simpler than maintaining two event types with identical payloads.

### 3. EventBus migration scoped to app-level cross-domain callbacks only
9 SetXxxCallback sites in wiring files are migrated. Domain-internal hooks (negotiation.SetEventCallback, SessionStore.SetInvalidationCallback) are explicitly excluded because they have different lifecycle and ownership — they're intra-domain, not cross-domain integration.

### 4. economy.BuildTools() receives individual engine pointers (not a struct)
The pilot extracts `buildEconomyTools(ec *economyComponents)` to `economy.BuildTools(be, re, ne, ee, pe)` taking individual exported engine types. An alternative was defining an `EconomyDeps` struct, but individual parameters are more explicit and don't require a new type in the domain package.

### 5. p2p.ConfigReader interface defined in consumer package
Following Go convention, the `ConfigReader` interface is defined in `internal/p2p/config.go` (the consumer), not in `internal/config/`. Getter methods are added to `config.P2PConfig` as value receivers so it implicitly satisfies the interface without an explicit adapter.

### 6. TextGenerator in internal/llm/ (not internal/types/)
`internal/types/` is already a catch-all with callbacks, enums, and context types. A purpose-specific `internal/llm/` package makes the interface discoverable and prevents further growth of the types junk drawer.

## Risks / Trade-offs

- **[EventBus ordering]** Synchronous EventBus dispatch maintains callback ordering, but if EventBus ever becomes async, graph/embed ordering assumptions break → Mitigation: EventBus.Publish is documented as synchronous; any async change would require explicit review
- **[NeedsGraph coupling]** Publishers must correctly set NeedsGraph; a bug means silent graph omission or pollution → Mitigation: regression test `TestContentSavedEvent_NeedsGraph` validates all 3 knowledge store paths
- **[economy pilot scope]** buildOnChainEscrowTools and buildSentinelTools remain in app — mixed ownership during transition → Mitigation: clearly documented as excluded in proposal; subsequent units planned
- **[ConfigReader method explosion]** P2PConfig has many fields; each becomes a getter method → Mitigation: only 6 methods needed for node.go; contained to 6 one-line value receivers
