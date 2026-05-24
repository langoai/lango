## Why

`lango mcp get` still writes its table-style inspection output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and makes this CLI surface inconsistent with the other hardened inspection commands.

## What Changes

- route `lango mcp get` output through `cmd.OutOrStdout()`
- add command-level capture coverage for a disabled server fixture
- sync MCP CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for MCP server inspection
- keeps user-visible output unchanged while aligning with the rest of the CLI
