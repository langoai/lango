## Context

The prompt package already centralizes hidden-input prompts, yes/no confirmation, and visible line-entry prompts. Extension pack install/remove still keeps a tiny local wrapper only for two extra rules: reject non-terminal stdin with a command-specific guidance message, and treat EOF as denial instead of as a hard error.

## Goals / Non-Goals

**Goals:**
- Centralize the guarded confirmation wrapper in the prompt package
- Preserve extension CLI behavior exactly
- Add direct tests at the prompt helper layer

**Non-Goals:**
- Generalizing every non-interactive policy in the repo in this turn
- Changing extension install/remove exit semantics
- Refactoring unrelated CLI prompt flows

## Decisions

Add `prompt.ConfirmTTYIO(...)` as a small wrapper over `ConfirmIO(...)`.
Rationale: the underlying yes/no parsing remains in one place, while the wrapper captures the common policy needed by extension and future commands.

Keep the non-TTY guidance message supplied by the caller.
Rationale: the prompt layer should enforce the guard, but command-specific UX wording should remain at the call site.

## Risks / Trade-offs

- [Risk] Another helper could over-expand the prompt API surface. → Mitigation: keep the helper narrow, stream-based, and focused on existing duplicated behavior.
- [Trade-off] Only extension adopts the helper immediately. → Mitigation: centralizing it now prevents new local wrappers and provides a ready target for future adoption.
