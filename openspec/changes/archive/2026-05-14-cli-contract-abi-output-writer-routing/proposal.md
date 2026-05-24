## Why

`lango contract abi load` still writes text and JSON output directly to process stdout. That makes command output harder to capture in wrappers and tests.

## What Changes

- Route `lango contract abi load` text and JSON output through `cmd.OutOrStdout()`
- Add command-level regression coverage for both output modes
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the contract ABI inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing ABI parsing behavior
