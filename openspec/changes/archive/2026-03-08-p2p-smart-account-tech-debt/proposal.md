# Proposal: P2P + Smart Account Technical Debt Resolution

## Problem

Paymaster feedback analysis revealed a `SpendingHookClient` ABI mismatch, triggering a comprehensive audit. Three parallel audit agents discovered **55 issues** (CRITICAL 21, HIGH 12, MEDIUM 14, LOW 8) across 5 root causes:

1. **ABI-First development not applied** — Solidity changes not reflected in Go bindings
2. **Scaffold-First pattern** — Structures created without implementation logic
3. **Callback disconnection** — Setters exist but never called in wiring
4. **Cross-layer isolation** — SmartAccount components private, inaccessible to other layers
5. **Missing tests** — No integration tests to detect connection gaps

## Solution

Systematic resolution across 5 phases (16 work units):

- **Phase A (ABI/Encoding)**: Fix SessionValidator, SpendingHook, UserOp hash, Safe initializer, nonce management, ABI dedup
- **Phase B (Security)**: SQL injection, session key encryption, handshake approval, ZK witness
- **Phase C (Wiring)**: On-chain session registration, budget engine sync, P2P CardFn/Gossip/TeamInvoke, component accessors
- **Phase D (Stubs→Real)**: CLI real implementation, policy syncer, paymaster recovery
- **Phase E (Tests)**: SmartAccount E2E, P2P connection, cross-layer integration tests

## Scope

- 29 files modified/created
- +1,089 / -484 lines changed
- 22 new integration tests
- Zero new dependencies
