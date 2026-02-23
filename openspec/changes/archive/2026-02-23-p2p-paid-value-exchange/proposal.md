## Why

The P2P, A2A, and Payment systems are fully isolated, preventing agents from executing a "provide value → receive USDC payment" flow. ZK proof circuits (4) are defined but not wired, PricingInfo is declared but unused, and the protocol handler's executor callback is disconnected, meaning P2P tool invocations don't actually work.

## What Changes

- Add canonical USDC contract registry with per-chain addresses and on-chain verification
- Add EIP-3009 `transferWithAuthorization` builder for gasless USDC payments
- Add Owner Privacy Shield that hard-blocks owner PII from P2P responses
- Add per-peer DID reputation system with trust scoring (Ent-backed)
- Add Payment Gate between firewall and tool executor for paid tool invocations
- Extend P2P protocol with `price_query` and `tool_invoke_paid` message types
- Wire all 4 ZK circuits (wallet ownership, response attestation, balance range, agent capability)
- Wire executor callback so P2P tool invocations actually execute
- Add buyer-side methods (`QueryPrice`, `InvokeToolPaid`) to remote agent

## Capabilities

### New Capabilities
- `p2p-payment-gate`: Payment verification gate for P2P tool invocations using EIP-3009 pre-signed authorizations
- `p2p-owner-shield`: Hard-block layer preventing owner PII leakage through P2P responses
- `p2p-reputation`: Per-peer trust scoring based on exchange outcomes (success/failure/timeout)
- `usdc-registry`: Canonical USDC contract address registry with on-chain verification

### Modified Capabilities
- `p2p-protocol`: Add price_query and tool_invoke_paid request types, payment_required status
- `p2p-firewall`: Add reputation checking and owner shield integration
- `p2p-networking`: Wire ZK proofs into handshake and attestation, wire executor callback

## Impact

- New packages: `internal/payment/contracts/`, `internal/payment/eip3009/`, `internal/p2p/paygate/`, `internal/p2p/reputation/`, `internal/p2p/firewall/owner_shield.go`
- New Ent schema: `PeerReputation` with trust score tracking
- Modified: `internal/p2p/protocol/handler.go`, `internal/p2p/protocol/messages.go`, `internal/p2p/protocol/remote_agent.go`
- Modified: `internal/p2p/firewall/firewall.go` (reputation + owner shield)
- Modified: `internal/app/wiring.go` (full wiring), `internal/app/app.go` (executor callback)
- Modified: `internal/config/types.go` (pricing, owner protection, min trust score configs)
- Dependencies: go-ethereum (already present), gnark (already present)
