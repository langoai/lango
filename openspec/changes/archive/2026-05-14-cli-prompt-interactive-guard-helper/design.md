## Context

The prompt package already owns shared helpers for hidden passphrase input, visible line-entry prompts, yes/no confirmation, TTY-only confirmation, and reusable TTY-file guards. Security CLI commands still directly call `prompt.IsInteractive()` and return their own error strings, even though the policy itself belongs next to the rest of the prompt helpers.

## Goals / Non-Goals

**Goals:**
- Centralize the `IsInteractive` guard pattern in the prompt package
- Preserve every existing command-specific error string
- Add direct helper coverage

**Non-Goals:**
- Changing how `IsInteractive()` itself detects terminals
- Refactoring commands outside the current security CLI call sites
- Changing passphrase prompt behavior

## Decisions

Add `prompt.RequireInteractiveTerminal(message string) error`.
Rationale: it makes the policy reusable while keeping the caller in charge of the exact UX wording.

Do not remove `prompt.IsInteractive()`.
Rationale: some callers may still need the boolean form, and the helper should layer on top of it rather than replace it.

## Risks / Trade-offs

- [Trade-off] Another small helper expands the prompt API slightly. → Mitigation: the helper is thin and directly corresponds to an already-repeated pattern.
- [Risk] Callers might accidentally pass divergent error strings. → Mitigation: this change preserves current strings exactly and only centralizes the guard logic.
