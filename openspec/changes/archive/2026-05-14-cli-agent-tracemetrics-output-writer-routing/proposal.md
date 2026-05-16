## Why

`lango agent trace metrics` still writes table and JSON output directly to process stdout instead of the Cobra command writer. That leaves one of the main trace-inspection surfaces inconsistent with the already-hardened CLI commands.

## What Changes

- route `agent trace metrics` table and JSON output through `cmd.OutOrStdout()`
- add command-level writer capture tests backed by a seeded real turn-trace store
- sync trace metrics docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for per-agent performance inspection
- keeps metric semantics unchanged while aligning output routing with the rest of the CLI
