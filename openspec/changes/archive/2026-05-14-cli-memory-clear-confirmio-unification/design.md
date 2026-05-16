## Context

The memory CLI already exposes command-writer-aware output and tests for clear confirmation, but the parser itself still uses a local scanner and manual `y/yes` matching. The shared prompt helper now covers the same interaction pattern and is already used by adjacent destructive commands.

## Goals / Non-Goals

**Goals:**
- Route `memory clear` confirmation through the shared prompt helper
- Preserve the existing warning text and force bypass
- Keep test coverage command-level and deterministic

**Non-Goals:**
- Refactoring memory graph clear in this same turn
- Changing deletion semantics
- Adding new flags or output modes

## Decisions

Use `prompt.ConfirmIO(...)` after printing the existing warning line.
Rationale: the prompt layer should own the yes/no parsing, while the memory command keeps control of its domain-specific warning text.

## Risks / Trade-offs

- [Risk] Prompt punctuation changes from the helper could drift from docs. → Mitigation: keep the prompt label as `Continue?` so the resulting text stays `Continue? [y/N]: ` and update assertions/spec accordingly.
- [Trade-off] The warning line remains outside the helper. → Mitigation: it is command-specific context, not generic prompt logic.
