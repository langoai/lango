## Context

The payment system has complete blockchain infrastructure: `LocalWallet` reads an encrypted private key from `SecretsStore` (`wallet.privatekey`), signs transactions, and submits them via RPC. However, no mechanism exists to *generate* the initial key. The agent responds "wallet creation not supported" because no tool exposes key generation.

All primitives are already available:
- `crypto.GenerateKey()` from go-ethereum for ECDSA key generation
- `SecretsStore.Store()` for encrypted persistent storage
- `LocalWallet` that reads from `SecretsStore` on demand

## Goals / Non-Goals

**Goals:**
- Enable the agent to create a new blockchain wallet via a `payment_create_wallet` tool
- Prevent accidental overwrite of existing wallet keys
- Return only the public address; never expose the private key
- Integrate seamlessly with existing `LocalWallet` and payment tools

**Non-Goals:**
- Mnemonic/seed phrase generation (BIP-39) — out of scope for now
- Multi-wallet support (only one wallet per instance)
- Key import from external sources
- Key export or backup functionality

## Decisions

### 1. Wallet creation lives in the `wallet` package as `CreateWallet()`

**Rationale**: Key generation is a wallet-level concern, not a payment-level concern. The `wallet` package already owns key storage logic (`local_wallet.go`). A standalone function `CreateWallet(ctx, secrets) (address, error)` keeps the responsibility in the right layer.

**Alternative considered**: Adding `CreateWallet` to `payment.Service` — rejected because the Service is about transactions, not key lifecycle.

### 2. Existence check before creation

**Rationale**: Overwriting an existing private key means permanent loss of funds. `CreateWallet` calls `secrets.Get("wallet.privatekey")` first. If a key exists, return an error with the existing address so the agent can inform the user.

**Alternative considered**: Force-overwrite with a `--force` flag — rejected for safety; the agent should never silently destroy a funded wallet.

### 3. SafetyLevel Dangerous with approval

**Rationale**: Creating cryptographic key material is an irreversible action. Consistent with `payment_send` being Dangerous. The approval system already handles this well.

### 4. Pass `SecretsStore` to tool builder

**Rationale**: The `payment_create_wallet` tool needs `SecretsStore` to generate and store the key. Currently `BuildTools()` only receives `payment.Service` and `SpendingLimiter`. We add `*security.SecretsStore` as a parameter and also pass network config (chainID, rpcURL) for address derivation context.

**Alternative considered**: Adding a `CreateWallet` method to `payment.Service` that delegates to wallet — adds unnecessary indirection.

## Risks / Trade-offs

- **[Risk] Accidental key overwrite** → Mitigated by existence check; returns error with existing address if wallet already exists
- **[Risk] Private key exposure in logs** → Mitigated by immediate zeroing after storage; tool returns only public address
- **[Risk] Key generated but storage fails** → Key is lost (acceptable: no funds at risk on a brand-new address). Error is returned to agent.
- **[Trade-off] No mnemonic backup** → Users cannot recover wallet from seed phrase. Acceptable for v1; can add BIP-39 later.
