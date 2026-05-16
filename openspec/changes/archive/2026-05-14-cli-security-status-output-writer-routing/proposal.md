## Why

`lango security status` still renders table and JSON output directly to process stdout instead of the Cobra command writer. The current tests also rely on process-global stdout swapping, which is weaker and less composable than command-level writer capture.

## What Changes

- route `security status` table and JSON output through `cmd.OutOrStdout()`
- make the shared `renderStatus` helper writer-aware
- update security status tests to capture output via buffers instead of swapping process stdout
- sync security CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for security inspection
- removes process-global stdout dependence from status rendering tests
