## Why

`lango contract read` still writes validation output directly to process stdout and its informational runtime note directly to process stderr. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango contract read` validation payloads through `cmd.OutOrStdout()`
- Route the informational runtime note through `cmd.ErrOrStderr()`
- Add command-level regression coverage for text and JSON modes with split stdout/stderr capture
- Update docs and OpenSpec with the stream-routing contract

## Impact

- Makes the contract read inspection surface consistent with the CLI stream-routing hardening work
- Improves testability without changing ABI validation behavior
