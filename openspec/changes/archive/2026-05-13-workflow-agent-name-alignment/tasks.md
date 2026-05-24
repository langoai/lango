## 1. Workflow Name Alignment

- [x] 1.1 Accept current built-in teammate names in workflow validation.
- [x] 1.2 Preserve legacy workflow agent names for backward compatibility.
- [x] 1.3 Update parser tests to cover both current and legacy names.

## 2. Verification

- [x] 2.1 Run `go test ./internal/workflow -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.
- [x] 2.4 Validate and archive the OpenSpec change.
