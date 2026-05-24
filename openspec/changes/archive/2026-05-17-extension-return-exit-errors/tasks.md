## 1. Tests First

- [x] 1.1 Add focused tests for a shared structured CLI exit-code error helper.
- [x] 1.2 Add a `cmd/lango` test proving `runMain()` preserves structured CLI exit code 3 and prints the underlying message once.
- [x] 1.3 Update extension command tests to expect returned structured errors instead of panic-based `os.Exit` interception.
- [x] 1.4 Add a repository guard rejecting direct `os.Exit` usage in non-test `internal/cli` files.

## 2. Implementation

- [x] 2.1 Add the shared `internal/cli/cliexit` helper.
- [x] 2.2 Refactor `internal/cli/extension` to return structured exit errors and remove its `os.Exit` seam.
- [x] 2.3 Update `cmd/lango/runMain` to preserve structured CLI exit codes.

## 3. Verification

- [x] 3.1 Run focused tests for `internal/cli/cliexit`, `internal/cli/extension`, `cmd/lango`, and `internal/testutil`.
- [x] 3.2 Run `openspec validate extension-return-exit-errors --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after implementation is verified.
