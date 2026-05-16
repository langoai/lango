## Context

The config profile commands already route their observable I/O through Cobra command streams, but `delete` still owns a one-off confirmation parser. The repository now has a shared prompt helper with explicit stream seams and coverage, so `delete` is a straightforward remaining duplicate.

## Goals / Non-Goals

**Goals:**
- Route `config delete` confirmation through the shared prompt helper
- Preserve the existing prompt text and `--force` bypass
- Add command-level coverage for approve, deny, and force flows

**Non-Goals:**
- Changing delete exit semantics
- Adding a new non-TTY policy for config delete in this turn
- Refactoring the rest of `configcmd` beyond the delete confirmation path

## Decisions

Use `prompt.ConfirmIO(cmd.InOrStdin(), cmd.OutOrStdout(), ...)` inside `newDeleteCmd`.
Rationale: it keeps the prompt format consistent with the rest of the CLI while preserving the command-stream contract already documented for config profile management.

Add regression coverage at the command level instead of unit-testing a small wrapper.
Rationale: the user-facing contract is that the full Cobra command path reads from `cmd.InOrStdin()` and writes to `cmd.OutOrStdout()`.

## Risks / Trade-offs

- [Risk] Shared helper semantics accept `yes` in addition to `y`, which is slightly broader than the old inline parser. → Mitigation: this is backward-compatible and aligns with other confirmation flows already using the shared helper.
- [Trade-off] EOF denial now flows through the shared helper behavior instead of bespoke parsing. → Mitigation: tests will assert that the command still aborts cleanly without deleting.
