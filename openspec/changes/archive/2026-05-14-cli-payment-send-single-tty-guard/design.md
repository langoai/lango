## Context

The payment send command now already calls `prompt.RequireTTYInput(...)` before prompting. A second branch still checks `cmd.InOrStdin() == os.Stdin && !paymentInteractiveCheck()`, which is both narrower and redundant.

## Goals / Non-Goals

**Goals:**
- Make `prompt.RequireTTYInput(...)` the only non-interactive confirmation gate for payment send
- Remove the redundant seam from tests

**Non-Goals:**
- Changing payment confirmation copy
- Refactoring other payment commands
- Changing JSON or success output

## Decisions

Delete `paymentInteractiveCheck` and the extra `os.Stdin` branch.
Rationale: the shared TTY-input guard already enforces the command contract for both real terminals and explicit file-based inputs.

## Risks / Trade-offs

- [Risk] Removing the extra branch could hide a subtle environment dependency. → Mitigation: the command-level non-interactive pipe regression remains and covers the actual contract.
