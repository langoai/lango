## Why

`lango graph import` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. The public CLI docs also describe flags and confirmation text that no longer match the runtime behavior.

## What Changes

- route `graph import` text and JSON output through `cmd.OutOrStdout()`
- add command-level writer capture tests for empty and imported graph import flows
- sync graph CLI docs and specs with the actual import contract

## Impact

- improves automation compatibility and testability for graph import
- removes public documentation drift around import flags and confirmation output
