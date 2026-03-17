## Why

After v0.2.0, rapid feature additions accumulated `map[string]interface{}` overuse (~1,285 instances), hardcoded strings, and inconsistent error handling across tool handlers and domain packages. This refactoring improves type safety, code consistency, and maintainability without changing any external behavior.

## What Changes

- New `internal/toolparam/` package providing type-safe parameter extraction (`RequireString`, `OptionalInt`, etc.) and standardized response builders (`StatusResponse`, `ListResponse`)
- Migrate 12 `tools_*.go` files from inline `params["key"].(string)` casts to `toolparam` helpers
- Extract domain-specific sentinel errors (`ErrNotFunded`, `ErrWorkspaceNotFound`, `ErrWorkflowDisabled`)
- Convert `AlertMetadata` from `map[string]interface{}` to typed struct in sentinel package
- Extract hardcoded validation maps to `internal/config/constants.go`
- Extract contract method name strings to `internal/economy/escrow/hub/methods.go` constants
- Add `writeMethod`/`readMethod` helpers to hub client, reducing ~150 lines of boilerplate
- Extract shared `windowCounter` type from sentinel detectors, removing ~40 lines of duplication
- Add `TransactionType` constants to escrow types
- Remove "failed to" error prefixes per go-errors.md conventions

## Capabilities

### New Capabilities
- `toolparam-extraction`: Type-safe parameter extraction and response building for tool handlers

### Modified Capabilities
- `sentinel-errors`: AlertMetadata typed struct, windowCounter extraction, domain error sentinels
- `economy-escrow`: Hub client writeMethod/readMethod helpers, method name constants, TransactionType constants
- `config-system`: Validation maps extracted to exported package-level constants

## Impact

- **Code**: 23 files modified + 8 new files across `internal/app/`, `internal/toolparam/`, `internal/config/`, `internal/economy/escrow/`, `internal/p2p/workspace/`, `internal/cli/workflow/`
- **APIs**: No external API changes — purely internal refactoring
- **Tests**: All 103 test packages pass; new table-driven tests for toolparam package
- **Net effect**: 651 insertions, 742 deletions (net -91 lines)
