## Why

`lango p2p sandbox` subcommands still write status text, smoke-test progress, and cleanup confirmations directly to process stdout. That makes command output harder to capture in wrappers and tests.

## What Changes

- Route `lango p2p sandbox status`, `test`, and `cleanup` output through `cmd.OutOrStdout()`
- Add seams for runtime constructors so command-level tests do not depend on live Docker or subprocess execution
- Add regression coverage for status, smoke-test, and cleanup output capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the sandbox management surface consistent with the CLI writer-routing hardening work
- Improves testability without changing sandbox behavior
