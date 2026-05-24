## 1. Contract Sync

- [x] 1.1 Make `fs_list` schema treat `path` as optional.
- [x] 1.2 Add regressions for optional-path schema and current-directory default behavior.
- [x] 1.3 Update prompt/spec wording for the optional `path` contract.

## 2. Verification

- [x] 2.1 Run `go test ./internal/tools/filesystem -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `tool-filesystem` and `downstream-docs-sync` coverage for the optional `path` contract.
- [x] 3.2 Validate and archive the OpenSpec change.
