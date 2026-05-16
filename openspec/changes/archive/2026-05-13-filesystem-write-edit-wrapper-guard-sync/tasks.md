## 1. Wrapper Guard

- [x] 1.1 Tighten `fs_write` and `fs_edit` to enforce declared required inputs.
- [x] 1.2 Add regression coverage for the missing-parameter wrapper paths.
- [x] 1.3 Update prompt/spec wording for the stricter write/edit contract.

## 2. Verification

- [x] 2.1 Run `go test ./internal/tools/filesystem -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `tool-filesystem` and `downstream-docs-sync` coverage for the write/edit wrapper guard contract.
- [x] 3.2 Validate and archive the OpenSpec change.
