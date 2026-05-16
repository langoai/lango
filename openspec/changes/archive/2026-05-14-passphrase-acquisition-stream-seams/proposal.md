## Why

The passphrase acquisition code still relied on process-global stdin/stderr for two important branches: stdin-pipe reading and keyring warning output. The behavior was correct, but tests had to replace global process streams to cover it.

## What Changes

- Add a reader-based stdin helper for passphrase pipe input
- Add an internal acquisition helper that accepts injected stdin/stderr
- Update tests to stop swapping `os.Stdin` for the covered non-interactive branches
- Record the new testability contract in OpenSpec

## Impact

- Improves test determinism for a security-critical code path
- Reduces reliance on process-global stream mutation in tests
