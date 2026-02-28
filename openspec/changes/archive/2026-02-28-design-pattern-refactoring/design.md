## Context

The application's initialization, startup, and shutdown are managed through a monolithic `App` struct with 22+ fields, a 440-line `New()` function, and manual `if != nil` checks in `Start()`/`Stop()`. Cross-cutting concerns (learning observation, approval gating, browser recovery) are applied via order-dependent nested wrapping loops. The bootstrap process runs 7 sequential steps with ad-hoc `client.Close()` calls for error cleanup. Component communication uses 13+ setter-based callbacks that couple stores to wiring code.

## Goals / Non-Goals

**Goals:**
- Replace manual Start/Stop with priority-ordered lifecycle registry with rollback
- Replace nested tool wrapping with composable middleware chain
- Replace bootstrap's ad-hoc cleanup with phase-based pipeline
- Create module system foundation for future declarative initialization
- Create event bus foundation for future callback elimination

**Non-Goals:**
- Migrate all existing `init*` functions to appinit modules (future PR)
- Migrate all callbacks to event bus (future PR, dual-mode)
- Change external API or CLI behavior
- Modify business logic

## Decisions

### D1: Lifecycle Registry with Priority Constants
**Decision**: Use numeric Priority constants (Infra=100, Core=200, Buffer=300, Network=400, Automation=500) with stable sort.
**Rationale**: Numeric priorities allow inserting new levels between existing ones without renaming. Stable sort preserves registration order within same priority.
**Alternative**: String-based priority or explicit dependency graph — rejected for simplicity.

### D2: HTTP-style Middleware for Tools
**Decision**: `Middleware func(tool *agent.Tool, next HandlerFunc) HandlerFunc` with Chain/ChainAll.
**Rationale**: Proven pattern from net/http. First middleware = outermost, build from inside out. Middlewares can short-circuit (approval denial) or pass-through (learning observation).
**Alternative**: Decorator pattern with wrapping structs — rejected as less composable.

### D3: Phase Pipeline with Cleanup Stack
**Decision**: Each Phase has optional Cleanup function. On failure, completed phases' Cleanups run in reverse.
**Rationale**: Directly models the resource acquisition/release pattern. No need for defer or context — explicit cleanup callbacks.
**Alternative**: Context-based cleanup (like Go's testing.Cleanup) — rejected to avoid context dependency in bootstrap.

### D4: Module System with Topological Sort
**Decision**: Kahn's algorithm for dependency resolution via Provides/DependsOn string keys.
**Rationale**: Well-understood O(V+E) algorithm, produces clear error messages on cycles, naturally handles disabled modules.
**Alternative**: Explicit ordering (current approach) — this is what we're replacing.

### D5: Synchronous Event Bus
**Decision**: Synchronous Publish with RWMutex. Generic SubscribeTyped[T] for type safety.
**Rationale**: Matches existing synchronous callback behavior. Async delivery would change semantics. Generics provide type safety without reflection.
**Alternative**: Channel-based async bus — rejected as it changes timing semantics of existing callbacks.

## Risks / Trade-offs

- [Risk] Registry adds indirection to startup/shutdown flow → Mitigated by clear logging in each component's Start/Stop
- [Risk] Middleware chain obscures tool wrapping order → Mitigated by explicit ordering in ChainAll call site
- [Risk] Event bus dual-mode migration period adds complexity → Mitigated by keeping old callbacks until all consumers migrate
- [Trade-off] Module system infrastructure added before migration — accepted as foundation for incremental adoption
