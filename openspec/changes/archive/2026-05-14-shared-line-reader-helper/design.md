## Context

Three separate packages currently use `bufio.NewReader(...).ReadString('\n')` for line-oriented input:
- `internal/cli/prompt` for CLI prompt helpers
- `internal/approval` for TTY approval fallback
- `internal/security/passphrase` for stdin-pipe acquisition

Each caller still needs its own domain-specific normalization and validation, but the raw line read itself is the same concern.

## Goals / Non-Goals

**Goals:**
- Centralize raw line reading in a lower-level shared package
- Preserve caller-specific behavior and error handling
- Keep the helper generic and dependency-light

**Non-Goals:**
- Unifying higher-level yes/no parsing or passphrase trimming semantics
- Reworking terminal detection in this turn
- Changing any public CLI prompt text

## Decisions

Create `internal/lineio` with a single `ReadLine(io.Reader) (string, error)` helper.
Rationale: this keeps the shared concern minimal and suitable for both CLI and non-CLI layers.

Keep caller-specific normalization outside the helper.
Rationale: approvals, confirmation prompts, and passphrase stdin parsing each need different follow-up behavior.

## Risks / Trade-offs

- [Trade-off] A new small package is introduced for a narrow helper. → Mitigation: it removes duplication across multiple layers and avoids inappropriate dependencies on `internal/cli/prompt`.
- [Risk] A raw helper could be overused for unrelated parsing. → Mitigation: keep the API intentionally tiny and line-oriented only.
