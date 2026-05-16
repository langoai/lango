## Context

The prompt package already owns shared helpers for hidden passphrase input, yes/no confirmation, visible line-entry prompts, and TTY-guarded confirmation. Three command paths still repeat the lower-level `*os.File` + `term.IsTerminal(...)` guard before using those helpers: payment send, keyring clear, and secrets delete.

## Goals / Non-Goals

**Goals:**
- Centralize the reusable TTY-input guard in the prompt package
- Reuse it from existing command paths without changing user-visible behavior
- Add direct helper coverage so command tests do not need to rediscover the same guard contract

**Non-Goals:**
- Changing prompt wording or command exit semantics
- Refactoring interactive-only passphrase entry guards in this turn
- Broadly redesigning the prompt API

## Decisions

Add `prompt.RequireTTYInput(...)` that returns nil for non-file readers and a caller-supplied error for non-terminal `*os.File` readers.
Rationale: commands often need the guard even when they do not want yes/no confirmation semantics.

Make `ConfirmTTYIO(...)` reuse `RequireTTYInput(...)`.
Rationale: this keeps the TTY guard logic defined in one place.

## Risks / Trade-offs

- [Risk] A too-generic helper could hide command intent. → Mitigation: keep the helper name explicit and leave the guidance message at the call site.
- [Trade-off] Only a subset of commands adopt it immediately. → Mitigation: centralizing it now still removes current duplication and provides the standard path for later adoption.
