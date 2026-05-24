## Why

`lango settings` still writes its cancel message and post-save guidance directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route settings cancel and completion output through `cmd.OutOrStdout()`
- Make the next-steps helper writer-aware
- Add writer-based tests and update docs/OpenSpec

## Impact

- Makes settings completion output consistent with the CLI writer-routing hardening work
- Avoids any future need for process-global stdout capture in settings tests
