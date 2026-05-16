## Context

The security CLI already routes confirmation prompts and success/warning output through Cobra command streams in several places, but `keyring clear` still keeps its own prompt rendering and input parsing logic. The shared prompt helper now covers the same semantics and is already used in adjacent security flows.

## Goals / Non-Goals

**Goals:**
- Route `keyring clear` confirmation through the shared prompt helper
- Preserve the existing non-interactive refusal unless `--force` is supplied
- Strengthen regression coverage around deny/confirm/force/non-interactive paths

**Non-Goals:**
- Changing the prompt text
- Changing warning or success output semantics after confirmation
- Refactoring other security subcommands in this turn

## Decisions

Use `prompt.ConfirmIO(cmd.InOrStdin(), cmd.OutOrStdout(), ...)` after the existing non-interactive guard.
Rationale: the guard is command-specific UX, while prompt formatting/parsing should live in the shared helper.

Keep the current `use --force for non-interactive deletion` message exactly as-is.
Rationale: the message is already tested and documented, and callers may rely on that exact guidance.

## Risks / Trade-offs

- [Risk] EOF denial behavior now comes through the shared helper instead of the local parser. → Mitigation: retain the existing `Aborted.` user-facing outcome in command-level tests.
- [Trade-off] A tiny command-local guard still remains. → Mitigation: keep only the non-interactive policy local and move all prompt I/O semantics to the shared helper.
