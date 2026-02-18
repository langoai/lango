# X402 V2 Protocol Migration

## Problem
The existing X402 payment implementation uses a custom V1-style protocol (custom `X-Payment-*` headers, direct ERC-20 `transfer()`) that is incompatible with the actual X402 V2 specification. The X402 Interceptor was never wired into the agent's HTTP flow.

## Solution
Migrate to the official Coinbase X402 V2 Go SDK (`github.com/coinbase/x402/go`) so agents can make automatic payments when encountering HTTP 402 responses from X402-enabled services.

### Key Changes
- Replace custom V1 protocol types with SDK's `PaymentRoundTripper`
- Use EIP-3009 `transferWithAuthorization` (off-chain signature) instead of direct `transfer()`
- Use CAIP-2 network identifiers (`eip155:84532`) instead of custom strings
- Add `x402_fetch` agent tool for HTTP requests with automatic payment
- Add `payment_method` enum to PaymentTx for distinguishing direct transfers from X402 auto-payments
- Wire X402 interceptor into the application lifecycle

## Scope
- Core: `internal/x402/` (rewrite), `internal/payment/` (modify)
- Application: `internal/tools/payment/` (add x402_fetch), `internal/app/` (wire)
- Schema: `internal/ent/schema/payment_tx.go` (add payment_method)
