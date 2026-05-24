## Context

The prompt package already owns hidden passphrase input and yes/no confirmation input, both with deterministic stream seams. Recovery mnemonic setup still uses a command-local `bufio.Reader` path for confirmation-word entry, which is the last prompt-shaped parser in that flow living outside the shared prompt layer.

## Goals / Non-Goals

**Goals:**
- Introduce a shared helper for visible line-entry prompts
- Route recovery confirmation-word entry through that helper
- Preserve the current prompt wording and mismatch behavior

**Non-Goals:**
- Changing mnemonic generation or recovery semantics
- Converting arbitrary free-form CLI input across the whole repo in this turn
- Changing passphrase hidden-input behavior

## Decisions

Add a `prompt.ReadLineIO(...)` helper that writes a prompt, reads one line, and returns the raw line text.
Rationale: it centralizes the stream-driven prompt/read behavior without forcing a yes/no or hidden-input semantic on callers.

Keep validation logic for recovery confirmation words in `confirmWord(...)`.
Rationale: the shared helper should only handle prompt I/O; recovery-specific word matching should remain in the recovery command.

## Risks / Trade-offs

- [Risk] A too-generic prompt helper could invite misuse. → Mitigation: keep the helper small, explicit, and stream-based, and only adopt it for the existing recovery line prompt in this turn.
- [Trade-off] Recovery still owns word normalization/mismatch checks. → Mitigation: that logic is domain-specific and belongs there.
