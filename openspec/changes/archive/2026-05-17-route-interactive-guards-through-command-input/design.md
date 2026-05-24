# Design: Command-Input-Aware Interactive Guards

## Boundary

The shared prompt package owns terminal/input guard semantics. CLI packages must
pass their Cobra command input stream into that helper before launching TUI flows
or hidden prompts.

## Approach

Add a small helper in `internal/cli/prompt` that accepts `io.Reader` plus a
caller-supplied guidance message. The helper delegates to the existing
`RequireTTYInput` behavior so non-terminal `*os.File` inputs are rejected while
non-file readers remain valid for tests and controlled injected-input flows.

Commands that currently call package-level `RequireInteractiveTerminal(...)`
will switch to package-local seams that accept `io.Reader`. This keeps tests
able to replace guard behavior where setup would otherwise require a real TTY,
without forcing process-global stdin mutation.

## Non-Goals

- Do not change prompt text, passphrase handling, encryption behavior, or command
  success output.
- Do not remove the existing `RequireInteractiveTerminal(...)` helper; it remains
  available for code paths that intentionally inspect process stdin.
- Do not broaden this change to every confirmation prompt. Confirmation commands
  that already use `RequireTTYInput(cmd.InOrStdin(), ...)` stay unchanged.

## Verification

- Add failing tests that prove onboard/settings and security interactive guards
  receive command input streams.
- Add a shared prompt helper test for injected readers and non-terminal file
  rejection.
- Run focused CLI prompt/security/onboard/settings tests, then full build/test
  and OpenSpec validation.
