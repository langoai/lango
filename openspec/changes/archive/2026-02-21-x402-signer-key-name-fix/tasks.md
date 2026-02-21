## 1. Fix Key Name

- [x] 1.1 Change `keyName` in `NewLocalSignerProvider()` from `"wallet_private_key"` to `"wallet.privatekey"` in `internal/x402/signer.go`

## 2. Verify

- [x] 2.1 Run `go build ./...` to confirm compilation
- [x] 2.2 Run `go test ./internal/x402/... ./internal/wallet/...` to confirm tests pass
