## 1. Wrapper Hardening

- [x] 1.1 Add a required string-slice helper for wrapper extraction.
- [x] 1.2 Make `session_key_create` reject missing `targets` and `duration`.
- [x] 1.3 Add exact missing-parameter regressions for the smart-account wrapper surface.

## 2. Docs And Spec Sync

- [x] 2.1 Update prompt/public docs for the smart-account required-input contract.
- [x] 2.2 Sync the smart-account and production-readiness specs to the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/toolparam ./internal/app -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
