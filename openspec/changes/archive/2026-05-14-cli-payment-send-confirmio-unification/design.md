## Context

The payment CLI already routes user-facing text through Cobra command writers, but `send` still owns a local yes/no parser even though adjacent commands now share a common confirmation helper. Because this path precedes a real payment submission, keeping confirmation semantics centralized is preferable.

## Goals / Non-Goals

**Goals:**
- Route `payment send` confirmation through the shared prompt helper
- Preserve the existing `--force` bypass and non-interactive refusal
- Keep the pre-confirmation payment summary text intact

**Non-Goals:**
- Changing payment business logic or receipt handling
- Changing JSON success output
- Refactoring other payment subcommands in this turn

## Decisions

Use `prompt.ConfirmIO(cmd.InOrStdin(), cmd.OutOrStdout(), "Confirm")` after printing the existing payment summary lines.
Rationale: this keeps the visible summary intact while consolidating the actual prompt formatting and parsing in one shared helper.

Keep the current `use --force for non-interactive mode` error unchanged.
Rationale: it is already part of the command's scripting contract and should remain stable.

## Risks / Trade-offs

- [Risk] The prompt text becomes `Confirm [y/N]: ` via the shared helper rather than manual formatting. → Mitigation: this matches the existing visible prompt, so no user-facing change occurs.
- [Trade-off] The summary remains split across separate `Fprintf` calls before confirmation. → Mitigation: this change intentionally targets the duplicate parser only.
