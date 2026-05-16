## 1. RPC Wallet Cleanup Symmetry

- [x] 1.1 Add missing `sign_tx` cleanup regressions across response/error/cancel paths.
- [x] 1.2 Add missing `sign_msg` cleanup regressions across response/error/timeout/cancel paths.

## 2. Verification

- [x] 2.1 Run `go test ./internal/wallet -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.
- [ ] 2.4 Validate and archive the OpenSpec change.
