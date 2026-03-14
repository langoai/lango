## Context

The current timeout system uses a fixed 5-minute `RequestTimeout` with an optional `AutoExtendTimeout` flag. Complex agent runs (multi-tool, long reasoning chains) regularly exceed 5 minutes while actively processing. When timeouts occur, the session history is left in an incomplete state (user message saved, assistant response missing), causing the next turn to receive garbled context.

## Goals / Non-Goals

**Goals:**
- Replace fixed timeout with idle-based timeout: requests stay alive while the agent produces activity
- Clean up session history on timeout to prevent error leakage into subsequent turns
- Full backward compatibility with existing `requestTimeout` and `autoExtendTimeout` configs
- Share the `ExtendableDeadline` type between `app` and `gateway` packages

**Non-Goals:**
- Client-side timeout control (client cancellation)
- Per-tool timeout changes (existing `toolTimeout` is unchanged)
- Changing the default behavior for existing installations (idle timeout is opt-in via config)

## Decisions

### Decision 1: Shared `deadline` package
**Choice**: Extract `ExtendableDeadline` to `internal/deadline/` with a backward-compat alias in `internal/app/`.
**Rationale**: Gateway needs the same idle timeout mechanism but cannot import `internal/app/`. A shared package avoids duplication. The alias preserves existing references without a large refactor.
**Alternative**: Duplicate the code in gateway — rejected due to maintenance burden.

### Decision 2: Idle timeout as opt-in config field
**Choice**: Add `IdleTimeout` field (default: 0 = disabled) rather than changing the default behavior.
**Rationale**: Existing users expect a fixed 5m timeout. Changing defaults would be a breaking behavioral change. New installs can set `idleTimeout: 2m` explicitly.
**Alternative**: Make idle timeout the default — rejected for backward compatibility.

### Decision 3: `resolveTimeouts()` precedence
**Choice**: 4-way config precedence in a single helper function:
1. `IdleTimeout > 0` → explicit idle mode
2. `IdleTimeout < 0` → explicitly disabled
3. `AutoExtendTimeout = true` → legacy compatibility mapping
4. Default → fixed timeout

**Rationale**: Single function encapsulates all timeout resolution logic, making it testable and preventing scattered conditionals.

### Decision 4: Session annotation on timeout
**Choice**: Add `AnnotateTimeout(key, partial)` to `session.Store` interface that appends a synthetic assistant message.
**Rationale**: The root cause of error leakage is an unpaired user message in history. Appending a synthetic assistant message closes the turn cleanly. Using the existing `AppendMessage` internally minimizes new code.
**Alternative**: Delete the incomplete turn — rejected because partial results are valuable context.

### Decision 5: `Reason()` method on `ExtendableDeadline`
**Choice**: Track why the deadline expired (`idle`, `max_timeout`, `cancelled`) via an enum.
**Rationale**: Callers need to distinguish idle timeout (E006) from hard ceiling (E001) for error classification and user messaging.

## Risks / Trade-offs

- **[Risk] Interface change to session.Store** → All mock implementations must be updated. Mitigated by adding no-op stubs to all existing mocks.
- **[Risk] Timer leak if Stop() not called** → Mitigated by always using `defer cancel()` pattern and storing `maxTimer` for explicit cleanup.
- **[Trade-off] Opt-in vs default** → New users don't get idle timeout automatically. Acceptable because existing behavior must not change.
