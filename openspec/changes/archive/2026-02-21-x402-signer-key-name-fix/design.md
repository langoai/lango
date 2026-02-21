## Context

The wallet package stores private keys under `"wallet.privatekey"` in SecretsStore (established in the blockchain-wallet spec). The x402 `LocalSignerProvider` was implemented with a different key name `"wallet_private_key"`, causing a silent lookup failure when attempting x402 payments after wallet creation.

## Goals / Non-Goals

**Goals:**
- Align the x402 signer's SecretsStore key name with the wallet package convention (`"wallet.privatekey"`)

**Non-Goals:**
- Refactoring the key name into a shared constant (single one-line fix is sufficient)
- Changing any wallet creation or storage logic

## Decisions

**Use `"wallet.privatekey"` as the canonical key name.**
Rationale: The wallet package defined this convention first (blockchain-wallet spec). The x402 signer was added later and used an inconsistent name. Changing the signer to match the wallet is the minimal, non-breaking fix since no production wallets were stored under `"wallet_private_key"`.

## Risks / Trade-offs

- [Risk] If any environment stored keys under `"wallet_private_key"` → Those keys would become unreachable. Mitigation: This key name was never reachable in practice (wallet creation always used `"wallet.privatekey"`), so no real data exists under the old name.
