## Context

The prompt package already distinguishes between raw confirmation (`ConfirmIO(...)`) and safer wrappers (`ConfirmDenyOnEOFIO(...)`, `ConfirmTTYIO(...)`). The top-level convenience wrapper still points at the raw helper, which is inconsistent with the package's safer-default direction.

## Goals / Non-Goals

**Goals:**
- Make the default confirmation wrapper safe-by-default on EOF
- Keep the raw helper available for explicit low-level callers

**Non-Goals:**
- Changing prompt punctuation or yes/no parsing
- Refactoring external command call sites in this turn

## Decisions

Route `Confirm(...)` through `ConfirmDenyOnEOFIO(...)`.
Rationale: the top-level wrapper should embody the safest general behavior, while raw `ConfirmIO(...)` remains opt-in for callers that explicitly want direct error semantics.

## Risks / Trade-offs

- [Trade-off] The existing read-error regression for `Confirm(...)` needs to split EOF-deny from true read errors. → Mitigation: keep the EOF case as denial and leave raw-error coverage on `ConfirmIO(...)`.
