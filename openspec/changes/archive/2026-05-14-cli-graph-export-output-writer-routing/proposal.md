## Why

`lango graph export` still wrote JSON and CSV output directly to process stdout instead of the Cobra command writer. The public CLI docs also described a file-path argument and `ntriples` format that do not match the runtime behavior.

## What Changes

- route `graph export` JSON and CSV output through `cmd.OutOrStdout()`
- add command-level writer capture tests for representative export flows
- sync graph CLI docs and specs with the actual export contract

## Impact

- improves automation compatibility and testability for graph export
- removes public documentation drift around export arguments and formats
