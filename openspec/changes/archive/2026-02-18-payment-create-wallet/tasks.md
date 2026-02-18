## 1. Wallet Creation Function

- [x] 1.1 Add `CreateWallet(ctx, secrets) (string, error)` to `internal/wallet/` that generates ECDSA key, stores encrypted in SecretsStore under `wallet.privatekey`, and returns public address
- [x] 1.2 Add existence check: if `wallet.privatekey` already exists, return `ErrWalletExists` with existing address
- [x] 1.3 Ensure private key bytes are zeroed immediately after storage

## 2. Payment Create Wallet Tool

- [x] 2.1 Add `buildCreateWalletTool(secrets, chainID)` in `internal/tools/payment/payment.go` returning `payment_create_wallet` tool with SafetyLevel Dangerous
- [x] 2.2 Tool returns `{address, chainId, network, status}` — status is `"created"` or `"exists"`
- [x] 2.3 Update `BuildTools()` signature to accept `*security.SecretsStore` and network config

## 3. Wiring

- [x] 3.1 Update `buildPaymentTools()` in `internal/app/tools.go` to pass SecretsStore and chainID
- [x] 3.2 Update `initPayment()` in `internal/app/wiring.go` to expose SecretsStore in `paymentComponents`

## 4. Verification

- [x] 4.1 Run `go build ./...` to verify compilation
- [x] 4.2 Run `go test ./...` to verify existing tests pass
