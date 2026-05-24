## Context

`prompt.ConfirmDenyOnEOFIO(...)` now exists and is already used by config delete, memory clear, graph clear, recovery written-down confirmation, and dead-letter retry. Payment send, keyring clear, and secrets delete still call `ConfirmIO(...)` directly, even though a missing confirmation input should also mean “no” there.

## Goals / Non-Goals

**Goals:**
- Apply the shared EOF-deny wrapper to the remaining destructive confirmation flows
- Keep all prompt text and non-interactive guard behavior unchanged

**Non-Goals:**
- Changing `ConfirmIO(...)` itself
- Refactoring non-confirmation prompt flows
- Changing any success/error text other than EOF no longer surfacing as a hard error

## Decisions

Replace direct `ConfirmIO(...)` calls with `ConfirmDenyOnEOFIO(...)` only in commands where denial is already the safe default.
Rationale: it preserves the same explicit-yes semantics while hardening missing-input behavior.

## Risks / Trade-offs

- [Trade-off] A few command tests need new EOF-abort expectations. → Mitigation: add focused command-level regressions to lock the behavior down.
