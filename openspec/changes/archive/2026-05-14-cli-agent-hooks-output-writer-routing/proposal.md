## Why

`lango agent hooks` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. That makes wrapper capture inconsistent with the hardened inspection commands and forces process-global stdout interception in tests.

## What Changes

- route `agent hooks` table/text and JSON output through `cmd.OutOrStdout()`
- update hook command tests to capture the Cobra output writer directly
- add explicit text-path writer capture coverage
- sync hooks CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for hook inspection
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
