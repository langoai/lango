## Context

`handleIdleKey()` already implements `Ctrl+C` as a warning on first press and quit on a second press within 500ms, with `Ctrl+D` as the immediate quit path. The mismatch lives in the visible help and public documentation rather than the control flow itself.

## Goals / Non-Goals

**Goals:**

- Make chat help copy and public docs describe idle/failed `Ctrl+C` honestly.
- Preserve streaming-state `Ctrl+C` cancellation wording.

**Non-Goals:**

- Change the underlying `Ctrl+C` quit behavior.
- Change approval or streaming controls.

## Decisions

- Use compact `quit x2` wording in the help bar.
  - Rationale: it keeps the narrow help surface readable while still communicating the double-press contract.

- Use fuller prose in `/help` and public docs.
  - Rationale: those surfaces can afford a clearer explanation of the first-press warning and `Ctrl+D` immediate-exit fallback.

## Risks / Trade-offs

- [Abbreviated `quit x2` wording may be slightly terse] → Reinforce the same semantics in `/help` and docs so operators are not left guessing.
