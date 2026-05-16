## Why

`lango security kms test` still writes roundtrip progress and success output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango security kms test` progress and success output through `cmd.OutOrStdout()`
- Add a small KMS provider seam for command-level testing
- Add regression coverage for the roundtrip output
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes `kms test` consistent with the rest of the KMS CLI writer-routing work
- Improves testability without changing KMS roundtrip behavior
