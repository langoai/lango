## 1. Tests First

- [x] 1.1 Add a failing `cliboot.BootResult` test proving it enables `StartStorageBroker`.
- [x] 1.2 Add a failing `cliboot.Config` test proving it enables `StartStorageBroker` and closes the result.
- [x] 1.3 Run the focused tests and confirm they fail before implementation.

## 2. Implementation

- [x] 2.1 Add a narrow `bootstrapRun` test seam in `internal/cli/cliboot`.
- [x] 2.2 Set `StartStorageBroker: true` in `BootResult`.
- [x] 2.3 Set `StartStorageBroker: true` in `Config` without changing close behavior.

## 3. Verification

- [x] 3.1 Run focused `cliboot` tests.
- [x] 3.2 Run `openspec validate enable-cliboot-storage-broker --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Run subagent review.
- [x] 3.5 Sync and archive the OpenSpec change.
