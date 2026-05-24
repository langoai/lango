## Why

`lango agent graph` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. That leaves the delegation-graph inspection surface inconsistent with the hardened `agent status`, `list`, and `trace` commands.

## What Changes

- route `agent graph` text and JSON output through `cmd.OutOrStdout()`
- add command-level writer capture tests backed by a seeded real trace store
- sync agent CLI docs and specs with the graph output-writer contract

## Impact

- improves automation compatibility and testability for delegation graph inspection
- keeps graph semantics unchanged while aligning output routing with the rest of the CLI
