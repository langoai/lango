## Why

`lango mcp add` and `lango mcp remove` still print confirmation output directly to process stdout instead of the Cobra command writer. That makes wrapper capture inconsistent with the already-hardened MCP inspection and lifecycle commands.

## What Changes

- route `mcp add` confirmation output through `cmd.OutOrStdout()`
- route `mcp remove` confirmation output through `cmd.OutOrStdout()`
- add command-level writer capture tests for project-scope add/remove flows
- sync MCP docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for MCP config mutation commands
- keeps user-visible messaging unchanged while aligning behavior with the rest of the CLI
