## Why

The P2P payment gate applies identical prepay conditions to all peers regardless of trust history, and the `SubmitOnChain()` method is a placeholder that never executes real settlements. This means high-reputation peers endure unnecessary friction, and successful payment authorizations are never settled on-chain. Completing these two gaps creates a trust-based dynamic payment pipeline with automated settlement.

## What Changes

- **Trust-weighted payment tiers**: `Gate.Check()` now routes peers by reputation score — high trust (>0.8) gets post-pay (tool first, settle later), medium trust stays prepay (EIP-3009 first), low trust is firewall-blocked as before.
- **Deferred payment ledger**: In-memory ledger tracks post-pay obligations with Add/Settle/Pending lifecycle.
- **Event-driven settlement**: New `ToolExecutionPaidEvent` triggers asynchronous on-chain `transferWithAuthorization` submission via a dedicated settlement service.
- **Settlement service**: Full lifecycle — DB record, EIP-1559 tx build, wallet signing, retry with exponential backoff, receipt polling with confirmation.
- **Reputation feedback loop**: Settlement success/failure feeds back into peer reputation scores.
- **`SubmitOnChain()` removed**: Replaced by the settlement service subscribed to the event bus.
- Config extended with `TrustThresholds` and `SettlementConfig` under `P2PPricingConfig`.
- Ent schema `payment_method` enum extended with `p2p_settlement`.

## Capabilities

### New Capabilities
- `trust-payment-tiers`: Reputation-based payment routing (post-pay vs prepay) in the payment gate
- `p2p-settlement`: Automated on-chain settlement service with event-driven architecture

### Modified Capabilities
- `p2p-pricing`: Extended config with trust thresholds and settlement parameters; removed `SubmitOnChain()`

## Impact

- **Core packages**: `internal/p2p/paygate/`, `internal/p2p/settlement/` (new), `internal/p2p/protocol/`, `internal/eventbus/`, `internal/config/`, `internal/app/`
- **Ent schema**: `payment_tx.go` payment_method enum change (requires `go generate`)
- **Dependencies**: Uses existing `eventbus.Bus`, `wallet.WalletProvider`, `ethclient.Client`, `reputation.Store`
- **Breaking**: `Gate.SubmitOnChain()` removed — callers should use the settlement service via event bus instead
