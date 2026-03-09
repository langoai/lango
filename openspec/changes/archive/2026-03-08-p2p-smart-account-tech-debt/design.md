# Design: P2P + Smart Account Technical Debt Resolution

## Architecture Decisions

### AD-1: ABI as source of truth
All Go bindings must exactly mirror the Solidity contract ABI. A `scripts/check-abi.sh` script validates this via `forge inspect`. This prevents future drift.

### AD-2: Packed UserOperation for v0.7
The `computeUserOpHash()` function packs gas fields into 2 words (`accountGasLimits`, `gasFees`) per the ERC-4337 v0.7 PackedUserOperation format. Go-side fields remain unpacked for readability; packing occurs only at hash computation.

### AD-3: Callback closure pattern for late binding
The P2P approval function captures a `*reputation.Store` pointer that is nil at creation time and backfilled later. This avoids restructuring the initialization order while maintaining the default-deny security posture.

### AD-4: RecoverableProvider wrapper
Rather than modifying each paymaster provider, a `RecoverableProvider` wraps any `PaymasterProvider` with retry and fallback logic. This preserves the provider interface and allows configuration via `FallbackMode`.

### AD-5: CLI deps initialization
CLI commands use a `smartAccountDeps` struct initialized from `bootstrap.Result`, following the established `initPaymentDeps` pattern. This avoids requiring a full `App` instance for CLI operations.

### AD-6: PolicySyncer for drift detection
A `PolicySyncer` bridges Go-side `HarnessPolicy` with on-chain `SpendingConfig`. It supports push, pull, and drift detection. Field mapping: MaxTxAmount→perTxLimit, DailyLimit→dailyLimit, MonthlyLimit→cumulativeLimit.

## Component Interactions

```
                    ┌─────────────────┐
                    │   CLI Commands   │
                    └────────┬────────┘
                             │ smartAccountDeps
                    ┌────────▼────────┐
                    │   App.SA Comps   │◄──── Accessors (C4)
                    └────────┬────────┘
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────┐ ┌──────▼──────┐ ┌─────▼─────┐
     │  Session Mgr │ │ Policy Eng  │ │  Manager   │
     │  (on-chain)  │ │ (+ Syncer)  │ │ (v0.7 hash)│
     └──────┬───────┘ └──────┬──────┘ └─────┬─────┘
            │                │              │
     ┌──────▼───────┐ ┌─────▼──────┐ ┌─────▼────────┐
     │ SV Binding   │ │ SH Binding │ │ Recoverable  │
     │ (corrected)  │ │ (rewritten)│ │ Paymaster    │
     └──────────────┘ └────────────┘ └──────────────┘
```
