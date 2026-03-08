# Design: ERC-7579 Smart Account

## Architecture

```
User (EOA / Master Key)
  │
  ├─ Owns Safe Smart Account (ERC-7579 via safe7579 adapter)
  │     ├─ LangoSessionValidator  (TYPE_VALIDATOR)
  │     ├─ LangoSpendingHook      (TYPE_HOOK)
  │     └─ LangoEscrowExecutor    (TYPE_EXECUTOR)
  │
  └─ Grants Session Key to Agent
        └─ SessionPolicy {allowedTargets, allowedFunctions, spendLimit, validUntil}
```

## Key Decisions

1. **Safe + Safe7579 adapter** — No custom account contracts. Only custom modules.
2. **Dual enforcement** — Off-chain (Go) for fast rejection + on-chain (Solidity) for tamper-proof guarantees.
3. **Callback injection** — `internal/smartaccount/` never imports economy/risk/sentinel. All cross-package wiring via typed function callbacks in `wiring_smartaccount.go`.
4. **Hierarchical sessions** — Master (user-created) → Task (agent-created, policy ≤ master).
5. **External bundler** — UserOps submitted via JSON-RPC. Supports any ERC-4337 bundler.
6. **Graceful degradation** — If `smartAccount.enabled: false`, all existing custody-model flows unchanged.
7. **Session key storage** — Private keys encrypted via CryptoProvider. Only public keys go on-chain.

## Package Structure

```
internal/smartaccount/
├── types.go           # Core types & AccountManager interface
├── errors.go          # Sentinel errors
├── manager.go         # AccountManager implementation
├── factory.go         # Safe CREATE2 deployment
├── session/           # Session key lifecycle
│   ├── store.go       # Store interface + MemoryStore
│   ├── manager.go     # Create/Revoke/Sign with callbacks
│   └── crypto.go      # ECDSA key generation/serialization
├── policy/            # Off-chain policy engine
│   ├── types.go       # HarnessPolicy, SpendTracker
│   ├── engine.go      # Per-account policy management
│   └── validator.go   # Pre-flight validation
├── module/            # ERC-7579 module registry
│   ├── registry.go    # Register/List/Get descriptors
│   └── abi_encoder.go # installModule/uninstallModule encoding
├── bundler/           # Bundler JSON-RPC client
│   ├── client.go      # eth_sendUserOperation etc.
│   └── types.go       # UserOpResult, GasEstimate
└── bindings/          # Contract ABI bindings
    ├── session_validator.go
    ├── spending_hook.go
    ├── escrow_executor.go
    └── safe7579.go
```

## Integration Flow

```
Risk Engine ── callback ──→ PolicyAdapter.Recommend()
                               │
                    Policy Engine (pre-flight)
                               │
                    Session Manager (sign UserOp)
                               │
                    Account Manager (submit via bundler)
                               │
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                    ▼
    Budget Sync          EventBus             Sentinel Guard
   (on-chain→off)       (publish)          (emergency revoke)
```
