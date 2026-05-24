## Why

`lango agent trace list` and `lango agent trace show` still write human-readable and JSON output directly to process stdout instead of the Cobra command writer. The public docs also show a shorthand trace detail invocation that does not match the actual command shape.

## What Changes

- route `agent trace list` table/JSON output through `cmd.OutOrStdout()`
- route `agent trace show` detail/JSON output through the same writer
- add command-level writer capture tests backed by a seeded real trace store
- sync agent CLI docs and specs with the actual trace command contract

## Impact

- improves automation compatibility and testability for trace inspection
- removes command-shape drift in the docs while keeping trace semantics unchanged
