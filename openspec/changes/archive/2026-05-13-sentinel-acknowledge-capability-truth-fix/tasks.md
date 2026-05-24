## 1. Capability Fix

- [x] 1.1 Reclassify `sentinel_acknowledge` as a state-mutating management tool.
- [x] 1.2 Add a regression that locks the updated capability metadata.

## 2. Docs And Spec Sync

- [x] 2.1 Update prompt/feature docs to say `sentinel_acknowledge` is the dangerous alert-state mutation path.
- [x] 2.2 Update the sentinel spec to require non-read-only capability metadata for acknowledgement.

## 3. Verification

- [x] 3.1 Run `go test ./internal/economy/escrow/sentinel -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
