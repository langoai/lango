## Why

`lango security kms wrap` and `lango security kms detach` still write success confirmations and multi-slot guidance directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango security kms wrap` success output through `cmd.OutOrStdout()`
- Route `lango security kms detach` success output and multi-slot guidance through `cmd.OutOrStdout()`
- Add regression coverage for wrap success, detach success, and multi-slot guidance
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the remaining KMS management mutations consistent with the CLI writer-routing hardening work
- Improves testability without changing envelope semantics
