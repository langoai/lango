# Proposal: Security Review Fixes

## Why

Codex code review (10 rounds) identified security vulnerabilities and functional regressions across the Phase 0-7 security/crypto changes. Issues span handshake identity binding, gossip card verification, bootstrap credential management, provenance DID mismatch, and ZK escrow verifier trust. Fixes are additive hardening — no new features.

## What Changes

- **Gossip card signature verification on receive** — `handleMessage()` now calls `VerifyCardSignature()` before storing cards. Empty bundles rejected. DID↔bundle v2 hash binding enforced. LegacyDID match removed (unverifiable without `Proofs.Legacy`).
- **Handshake bundle cache timing** — Bundle cache and alias registration moved after authentication in both `HandleIncoming` and `Initiate`. v2 DID↔bundle hash + signing key binding added.
- **Handshake DID↔pubkey binding** — v1 DID↔pubkey consistency check added. v2 requires bundle with matching signing key. Alias registration deferred until after approval.
- **Bootstrap phase order** — `phaseLoadSecurityState` moved before `phaseMigrateEnvelope`. Pending migration/rekey loads salt even when envelope exists.
- **Status keyring + config** — `readDBStatusNonInteractive` passes keyring provider. Loads active config from DB when MK available. Keyfile fallback on stale keyring for both envelope and legacy paths.
- **Credential sync** — `change-passphrase` and `recovery restore` update keyfile and keyring after rotation. Keyring update always attempted (interactive command).
- **Provenance DID/signer alignment** — Provenance export uses wallet v1 DID (secp256k1) instead of v2 DID to match `VerifyMessageSignature` expectations.
- **Economy resolver** — `selectSettler` accepts `AddressResolver` for DID v2 settlement support.
- **ZK escrow verifier pinning** — `LangoZKEscrow` verifier address pinned as immutable in constructor, removed from `releaseWithProof` parameters.

## Capabilities

### Modified Capabilities
- `p2p-discovery` — gossip card signature verification, DID↔bundle binding
- `p2p-handshake` — bundle cache timing, DID↔pubkey binding, alias registration order
- `bootstrap-lifecycle` — phase order, pending migration salt loading
- `cli-security-status` — keyring provider, config loading, keyfile fallback
- `passphrase-management` — credential sync after rotation
- `onchain-escrow` — ZK escrow verifier pinning
