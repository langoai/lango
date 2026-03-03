## Why

The P2P payment system has a complete seller-side flow (paygate → trust-based routing → settlement → on-chain confirmation) and X402 HTTP 402 auto-intercept with SpendingLimiter, but the buyer side lacks an automated tool to invoke paid remote tools. Currently, `p2p_price_query` can check pricing, but there is no tool that combines price checking, EIP-3009 authorization signing, spending limit enforcement, and paid tool invocation into a single automated call. Buyers must manually construct payment authorizations, which breaks the seamless agent-to-agent interaction model.

## What Changes

- Add `p2p_invoke_paid` tool that automates the full buyer-side paid invocation flow: session check → price query → free/paid routing → spending limit check → EIP-3009 signing → `InvokeToolPaid()` call → response handling
- Add `authToMap()` helper that serializes `eip3009.Authorization` into the map format expected by the seller-side `paygate.parseAuthorization()`
- Wire `p2p_invoke_paid` into the P2P tool registration alongside existing `p2p_pay` and `p2p_price_query`
- Connect the existing `SpendingLimiter` (previously X402-only) to P2P buyer purchases

## Capabilities

### New Capabilities
- `p2p-buyer-auto-payment`: Buyer-side automatic payment tool that combines price query, spending limit enforcement, EIP-3009 authorization signing, and paid tool invocation into a single agent tool

### Modified Capabilities
- `p2p-payment`: Extended with buyer-side auto-payment tool registration alongside existing `p2p_pay`

## Impact

- `internal/app/tools_p2p.go`: New `buildP2PPaidInvokeTool()` function and `authToMap()` helper
- `internal/app/app.go`: Additional tool registration line in P2P wiring block
- Reuses existing infrastructure: `eip3009.Sign()`, `wallet.SpendingLimiter`, `protocol.InvokeToolPaid()`, `protocol.QueryPrice()`
- No breaking changes; additive only
