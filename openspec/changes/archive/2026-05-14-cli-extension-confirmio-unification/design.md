## Context

The extension CLI already uses Cobra command streams for inspect output and cancellation messages, but its interactive confirmation path still uses a local parser. The repository recently added seam-aware confirmation behavior in `internal/cli/prompt`, so the extension commands are now the clearest remaining duplicate in this area.

## Goals / Non-Goals

**Goals:**
- Route extension install/remove confirmation through the shared prompt helper
- Preserve the current non-TTY refusal behavior when `--yes` is omitted
- Add regression tests at the command level so wrappers can rely on the command streams

**Non-Goals:**
- Changing exit-code semantics
- Changing the trust model or inspect output shape
- Generalizing non-TTY confirmation policy for every CLI command in this turn

## Decisions

Keep a tiny extension-local guard for non-TTY stdin, but delegate prompt rendering and answer parsing to `prompt.ConfirmIO(...)`.
Rationale: the non-TTY error is extension-specific UX, while the actual prompt interaction rules should live in the shared helper.

Add command-level tests for install/remove confirmation rather than only unit-testing a small wrapper.
Rationale: the important contract is that the full Cobra command path uses `cmd.InOrStdin()` and `cmd.OutOrStdout()` correctly.

## Risks / Trade-offs

- [Risk] Shared helper integration could subtly change how empty input is treated. → Mitigation: add command-level denial coverage and keep the existing cancel messaging.
- [Trade-off] A small extension-local helper still remains for the non-TTY guard. → Mitigation: limit it to the guard and keep all prompt formatting/parsing in the shared prompt package.
