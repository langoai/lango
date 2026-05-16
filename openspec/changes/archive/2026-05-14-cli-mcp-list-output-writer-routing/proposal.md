## Why

`lango mcp list` still writes table and empty-state output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves MCP inspection inconsistent across subcommands.

## What Changes

- route `mcp list` table output through `cmd.OutOrStdout()`
- route the empty-state guidance through the same writer
- add command-level capture tests for configured and empty states
- sync MCP specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for MCP server listing
- keeps user-visible output unchanged while aligning with other CLI surfaces
