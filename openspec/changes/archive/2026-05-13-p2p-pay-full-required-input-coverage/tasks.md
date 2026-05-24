## 1. Regression Coverage

- [x] 1.1 Add exact missing-parameter regressions for `p2p_pay` required `peer_did` and `amount`.

## 2. Docs And Spec Sync

- [x] 2.1 Sync the production-readiness and P2P payment docs/spec wording to the full required-input set.

## 3. Verification

- [x] 3.1 Run `go test ./internal/app -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
