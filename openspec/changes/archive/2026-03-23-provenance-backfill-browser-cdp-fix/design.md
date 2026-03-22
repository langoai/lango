## Context

Two independent runtime bugs discovered during navigator agent execution. Neither is caused by recent ADK refactoring — both are pre-existing.

## Goals / Non-Goals

**Goals:**
- Provenance tree backfill for sessions loaded via Get() (missing-only, no status reset)
- CDP error recovery for browser_navigate with session reset retry

**Non-Goals:**
- Modifying browser_action retry behavior (side-effect risk)
- Changing provenance tree storage semantics

## Decisions

### D1: Idempotent rootObserver over EnsureRegistered API
**Decision**: Make the rootObserver closure idempotent (GetNode check before RegisterSession) rather than adding a new EnsureRegistered method.
**Rationale**: Keeps the change localized to wiring.go closure. No provenance API surface change needed.

### D2: Navigate-only CDP retry scope
**Decision**: CDP target error retry applies only to `browser_navigate`, not to any `browser_action` or other browser tools.
**Rationale**: browser_action includes click/type/eval with side effects. Auto-retry could cause duplicate form submissions or double-clicks.

## Risks / Trade-offs

- **[Risk] Idempotent check adds a GetNode call per Get()** → Negligible cost (in-memory map lookup). Only fires when rootSessionObserver is set (provenance enabled).
- **[Trade-off] Navigate retry creates a fresh session** → User loses any cookies/state from the old session. Acceptable for navigation which is the first step in a browser flow.
