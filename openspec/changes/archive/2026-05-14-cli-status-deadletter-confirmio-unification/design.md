## Context

The status CLI already sanitizes dead-letter IDs and routes confirmation text through Cobra command streams, but its retry confirmation still keeps a local parser. The shared prompt helper now covers the same interaction pattern and is used across adjacent config, security, extension, and payment flows.

## Goals / Non-Goals

**Goals:**
- Route dead-letter retry confirmation through the shared prompt helper
- Preserve the existing EOF-as-deny behavior
- Keep the sanitized receipt ID and prompt wording intact

**Non-Goals:**
- Changing retry semantics or actor injection
- Refactoring dead-letter list/detail output
- Introducing a non-interactive force flag for this command

## Decisions

Wrap `prompt.ConfirmIO(...)` with a tiny adapter that maps `io.EOF` to `(false, nil)`.
Rationale: this keeps the current "empty input means abort" behavior while consolidating prompt formatting/parsing in the shared helper.

Keep the existing sanitized receipt ID interpolation outside the shared helper.
Rationale: the sanitization is command-specific, while the prompt interaction rules should be shared.

## Risks / Trade-offs

- [Risk] Prompt output could drift if the command keeps manual formatting. → Mitigation: delegate the prompt suffix and answer parsing to the shared helper.
- [Trade-off] A small wrapper remains for EOF compatibility. → Mitigation: keep it local and single-purpose.
