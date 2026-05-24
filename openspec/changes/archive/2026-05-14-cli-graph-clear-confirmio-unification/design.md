## Context

The graph CLI already documents its command-stream confirmation contract, but `clear` still keeps a manual scanner and local yes/no parsing. The memory CLI just moved the same pattern onto the shared prompt helper, so graph clear is the next obvious duplicate.

## Goals / Non-Goals

**Goals:**
- Route `graph clear` confirmation through the shared prompt helper
- Preserve the existing warning line and force bypass
- Keep current tests command-level and deterministic

**Non-Goals:**
- Refactoring other graph commands in this turn
- Changing graph store behavior
- Adding new flags or output modes

## Decisions

Use `prompt.ConfirmIO(...)` after printing the existing warning line.
Rationale: the helper should own the yes/no parsing; the graph command should keep only the domain-specific warning text.

## Risks / Trade-offs

- [Risk] Prompt punctuation could drift from docs if the message changes. → Mitigation: keep the prompt label as `Continue?` so the resulting `Continue? [y/N]: ` remains stable.
- [Trade-off] The warning line remains separate from the helper. → Mitigation: it is command-specific context, not generic prompt logic.
