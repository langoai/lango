## Why

`lango mcp test` still writes its handshake and ping diagnostics directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves MCP diagnostics inconsistent across subcommands.

## What Changes

- route `mcp test` diagnostic output through `cmd.OutOrStdout()`
- add command-level capture coverage for a failing handshake fixture
- sync MCP specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for MCP connectivity diagnostics
- keeps user-visible output unchanged while aligning with other CLI surfaces
