# Design: Harden Passphrase Public Stdio Seams

## Approach

Add package-level variables in `internal/security/passphrase` for the process stdin reader, stderr writer, and terminal detection used by public wrappers. `Acquire` will pass those seams into the existing `acquireWithIO` helper, and `AcquireNonInteractive` will pass the stderr seam into `acquireNonInteractiveWithIO`.

## Test Strategy

Add non-parallel tests because they temporarily replace package-level seams. The tests will restore seams with `t.Cleanup` and verify:

- `Acquire` reads stdin from the injected reader when terminal detection is false.
- `Acquire` writes keyring warnings to the injected writer.
- `AcquireNonInteractive` writes keyring warnings to the injected writer.

## Compatibility

Default seam values remain `os.Stdin`, `os.Stderr`, and `term.IsTerminal(int(syscall.Stdin))`, preserving runtime behavior.
