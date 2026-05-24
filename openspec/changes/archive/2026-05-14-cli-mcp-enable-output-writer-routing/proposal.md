## Why

`lango mcp enable` and `lango mcp disable` still print their confirmation messages directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves MCP lifecycle operations inconsistent with the rest of the hardened CLI surfaces.

## What Changes

- route `mcp enable/disable` confirmation output through `cmd.OutOrStdout()`
- add project-scope command-level capture coverage for both actions
- sync MCP specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for MCP lifecycle mutations
- keeps user-visible output unchanged while aligning with other commands
