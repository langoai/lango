## Why

The x402 `LocalSignerProvider` uses `"wallet_private_key"` as the SecretsStore key name, while the wallet package (`CreateWallet`, `LocalWallet`) uses `"wallet.privatekey"`. This mismatch means that after creating a wallet, the x402 payment signer cannot find the private key, causing all x402 payments to fail with a "key not found" error.

## What Changes

- Fix the secret key name in `internal/x402/signer.go` from `"wallet_private_key"` to `"wallet.privatekey"` to match the wallet package convention

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `x402-v2`: Fix SecretsStore key name to align with the wallet package's `"wallet.privatekey"` convention

## Impact

- `internal/x402/signer.go`: Key name constant change (line 29)
- No API changes, no breaking changes
- Existing wallets created via `payment_create_wallet` tool will now be correctly loaded by the x402 signer
