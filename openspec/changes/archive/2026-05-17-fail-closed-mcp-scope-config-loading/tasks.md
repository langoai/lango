## 1. Tests First

- [x] 1.1 Add failing MCP loader tests for missing scoped files and invalid present scoped files.
- [x] 1.2 Add failing MCP CLI tests for invalid project config on read and write commands.

## 2. Implementation

- [x] 2.1 Make scoped MCP merge ignore missing files only and return actionable errors for invalid present files.
- [x] 2.2 Update MCP CLI read commands to surface scoped merge errors.
- [x] 2.3 Update MCP CLI write commands to avoid overwriting invalid existing scoped files.

## 3. Docs and Verification

- [x] 3.1 Update public MCP docs to document missing-vs-invalid scoped config behavior.
- [x] 3.2 Run focused MCP tests and OpenSpec validation.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Review, sync, and archive the OpenSpec change.
