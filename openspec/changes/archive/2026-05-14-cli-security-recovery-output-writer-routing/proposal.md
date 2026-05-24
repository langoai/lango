## Why

`lango security recovery setup` and `restore` still write their non-error output directly to process stdout, and setup's confirmation-word prompt also reads and writes via process-global streams. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route recovery setup mnemonic banner, confirmation-word prompt, and success output through command streams
- Route recovery restore success output through `cmd.OutOrStdout()`
- Add execution seams for deterministic command-level tests
- Update docs and OpenSpec with the recovery stream contract

## Impact

- Makes the recovery CLI surfaces consistent with the CLI writer-routing hardening work
- Improves testability without changing recovery semantics
