## Why

The core CLI reference now documents top-level stdout/stderr routing contracts for `serve`, `version`, `health`, and the interactive TUI entrypoints, but README still treats those commands only as feature bullets. That leaves the public summary docs slightly behind the current wrapper/test-harness reality.

## What Changes

- Add a short README note describing top-level utility stdout routing
- Add a short README note describing workbench/cockpit/chat startup stderr routing
- Record the README truth-sync expectation in the docs-only spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `docs-only`: README top-level command summaries stay aligned with current stream-routing contracts

## Impact

- Affected docs: `README.md`
- Affected specs: `openspec/specs/docs-only/spec.md`
- No code or runtime behavior changes
