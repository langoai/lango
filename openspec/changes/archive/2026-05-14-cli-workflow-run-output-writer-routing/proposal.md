## Why

`lango workflow run` still writes its validation summary, schedule guidance, and direct execution status output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route `lango workflow run` non-error output through `cmd.OutOrStdout()`
- Replace the schedule-path stdout swap test with writer-based capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes workflow run output consistent with the CLI writer-routing hardening work
- Removes process-global stdout swapping from workflow run tests
