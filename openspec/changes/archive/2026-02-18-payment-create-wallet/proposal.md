## Why

The payment system has full blockchain infrastructure (LocalWallet, SecretsStore, spending limits, USDC transfers) but lacks a wallet creation tool. When a user asks the agent to create a wallet, the agent cannot do so because no `payment_create_wallet` tool exists — even though all the underlying primitives (key generation via go-ethereum, encrypted storage via SecretsStore) are already available. This forces users to manually generate and import keys outside the agent, breaking the self-service experience.

## What Changes

- Add a `payment_create_wallet` agent tool that generates a new ECDSA private key, stores it encrypted in SecretsStore under `wallet.privatekey`, and returns only the public address
- The tool checks if a wallet already exists to prevent accidental key overwrite
- SafetyLevel is Dangerous (requires approval) since it creates cryptographic key material
- Add a `CreateWallet` method to the wallet package that encapsulates key generation + storage
- Wire the new tool into the payment tools builder so it's available to the Vault agent

## Capabilities

### New Capabilities

### Modified Capabilities
- `blockchain-wallet`: Add wallet creation/generation requirement to the existing wallet provider capability
- `payment-tools`: Add `payment_create_wallet` tool requirement to the existing payment tools capability

## Impact

- `internal/wallet/` — New `CreateWallet()` function for key generation + encrypted storage
- `internal/tools/payment/payment.go` — New `payment_create_wallet` tool builder
- `internal/app/wiring.go` — Pass SecretsStore to tool builder for wallet creation
- Vault agent gains wallet creation capability via existing `payment_` tool prefix routing
