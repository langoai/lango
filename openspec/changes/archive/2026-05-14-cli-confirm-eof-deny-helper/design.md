## Context

`prompt.ConfirmIO(...)` is intentionally low-level and currently returns read errors directly. Several command flows want a stricter UX rule: EOF or empty confirmation input should mean “deny” rather than “crash the command”. A few commands already re-implement that behavior locally.

## Goals / Non-Goals

**Goals:**
- Centralize EOF-as-deny confirmation behavior in the prompt package
- Reuse it from commands where denial is the safe default
- Keep raw `ConfirmIO(...)` available for callers that need direct error semantics

**Non-Goals:**
- Changing `ConfirmIO(...)` itself
- Changing TTY guard behavior
- Refactoring unrelated prompt flows

## Decisions

Add `prompt.ConfirmDenyOnEOFIO(...)` as a thin wrapper over `ConfirmIO(...)`.
Rationale: it makes the safer behavior explicit without mutating the lower-level helper's contract.

Use the wrapper only for flows where deny-on-missing-input is the intended behavior.
Rationale: callers should opt into the safer semantics deliberately.

## Risks / Trade-offs

- [Trade-off] The prompt API grows by one small wrapper. → Mitigation: it removes repeated EOF handling and makes the intended behavior explicit.
- [Risk] Some tests may implicitly rely on error behavior today. → Mitigation: update command-level regressions to assert the safer denial behavior directly.
