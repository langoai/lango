## 1. Helper Coverage

- [x] 1.1 Add direct tests for `sortedParamKeys`.
- [x] 1.2 Add direct tests for `formatParamValue`.
- [x] 1.3 Add direct tests for `singleLineValue`.
- [x] 1.4 Add direct tests for `compactRequestID`.

## 2. Spec Sync

- [x] 2.1 Record the helper-coverage requirement in the OpenSpec delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-helper-unit-coverage --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
