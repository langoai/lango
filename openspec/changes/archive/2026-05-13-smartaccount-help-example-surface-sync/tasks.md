## 1. CLI Help Sync

- [x] 1.1 Add a representative paymaster example to the top-level `lango account` help text.
- [x] 1.2 Add a CLI regression covering the updated account help output.

## 2. Docs And Spec Sync

- [x] 2.1 Update the smart account CLI docs to include the same top-level example surface.
- [x] 2.2 Update the smartaccount downstream spec to require paymaster coverage in the top-level CLI overview.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/smartaccount -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
